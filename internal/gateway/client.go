package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: timeout},
	}
}

type accountResponse struct {
	Data struct {
		Account struct {
			Nonce   uint64 `json:"nonce"`
			Balance string `json:"balance"`
		} `json:"account"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}

func (c *Client) GetAccount(ctx context.Context, bech32 string) (uint64, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/address/%s", c.baseURL, bech32), nil)
	if err != nil {
		return 0, "", err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	var out accountResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, "", err
	}
	if out.Error != "" && out.Error != "successful" {
		return 0, "", fmt.Errorf("gateway error: %s (%s)", out.Error, out.Code)
	}
	return out.Data.Account.Nonce, out.Data.Account.Balance, nil
}

func (c *Client) GetAccountNonce(ctx context.Context, bech32 string) (uint64, error) {
	nonce, _, err := c.GetAccount(ctx, bech32)
	if err != nil {
		return 0, err
	}
	return nonce, nil
}

func (c *Client) GetAccountBalance(ctx context.Context, bech32 string) (string, error) {
	_, balance, err := c.GetAccount(ctx, bech32)
	if err != nil {
		return "", err
	}
	return balance, nil
}

type TxSendRequest struct {
	Nonce    uint64 `json:"nonce"`
	Value    string `json:"value"`
	Receiver string `json:"receiver"`
	Sender   string `json:"sender"`
	GasPrice uint64 `json:"gasPrice"`
	GasLimit uint64 `json:"gasLimit"`
	Data     string `json:"data,omitempty"`
	ChainID  string `json:"chainID"`
	Version  uint32 `json:"version"`
	Options  uint32 `json:"options,omitempty"`
	Relayer  string `json:"relayer,omitempty"`

	Signature        string `json:"signature,omitempty"`
	RelayerSignature string `json:"relayerSignature,omitempty"`
}

type sendResponse struct {
	Data struct {
		TxHash string `json:"txHash"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}

type sendMultipleResponse struct {
	Data struct {
		NumOfSentTxs int               `json:"numOfSentTxs"`
		TxsHashes    map[string]string `json:"txsHashes"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}

func (c *Client) SendTx(ctx context.Context, tx TxSendRequest) (string, error) {
	b, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	var lastErr error
	backoff := 500 * time.Millisecond

	for attempt := 1; attempt <= 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/transaction/send", bytes.NewReader(b))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
			if attempt == 4 || !IsTransientSendError(err) {
				return "", err
			}
			if sleepErr := sleepWithContext(ctx, backoff); sleepErr != nil {
				return "", sleepErr
			}
			backoff *= 2
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			if attempt == 4 {
				return "", readErr
			}
			if sleepErr := sleepWithContext(ctx, backoff); sleepErr != nil {
				return "", sleepErr
			}
			backoff *= 2
			continue
		}

		var out sendResponse
		decodeErr := json.Unmarshal(body, &out)
		if decodeErr != nil {
			lastErr = decodeErr
			if attempt == 4 {
				return "", fmt.Errorf("send tx decode response: http=%d body=%s err=%w", resp.StatusCode, compactBody(body), decodeErr)
			}
			if sleepErr := sleepWithContext(ctx, backoff); sleepErr != nil {
				return "", sleepErr
			}
			backoff *= 2
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusBadGateway || resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusGatewayTimeout {
			lastErr = fmt.Errorf("gateway send returned HTTP %d body=%s", resp.StatusCode, compactBody(body))
			if attempt == 4 {
				return "", lastErr
			}
			if sleepErr := sleepWithContext(ctx, backoff); sleepErr != nil {
				return "", sleepErr
			}
			backoff *= 2
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return "", fmt.Errorf("gateway send returned HTTP %d error=%s code=%s body=%s", resp.StatusCode, out.Error, out.Code, compactBody(body))
		}

		if out.Error != "" && out.Error != "successful" {
			err := fmt.Errorf("gateway error: %s (%s) http=%d body=%s", out.Error, out.Code, resp.StatusCode, compactBody(body))
			if attempt == 4 || !IsTransientSendError(err) {
				return "", err
			}
			lastErr = err
			if sleepErr := sleepWithContext(ctx, backoff); sleepErr != nil {
				return "", sleepErr
			}
			backoff *= 2
			continue
		}
		return out.Data.TxHash, nil
	}

	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("send tx failed after retries")
}

func IsTransientSendError(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "deadline exceeded") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "temporarily unavailable") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "eof") ||
		strings.Contains(msg, "too many requests") ||
		strings.Contains(msg, "bad gateway") ||
		strings.Contains(msg, "service unavailable") ||
		strings.Contains(msg, "gateway timeout") ||
		strings.Contains(msg, "sending request error") ||
		strings.Contains(msg, "lowernonceintx") ||
		strings.Contains(msg, "veryhighnonceintx")
}

func (c *Client) SendTxs(ctx context.Context, txs []TxSendRequest) ([]string, error) {
	if len(txs) == 0 {
		return nil, nil
	}

	b, err := json.Marshal(txs)
	if err != nil {
		return nil, err
	}

	var lastErr error
	backoff := 500 * time.Millisecond

	for attempt := 1; attempt <= 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/transaction/send-multiple", bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
			if attempt == 4 || !IsTransientSendError(err) {
				return nil, err
			}
			if sleepErr := sleepWithContext(ctx, backoff); sleepErr != nil {
				return nil, sleepErr
			}
			backoff *= 2
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			if attempt == 4 {
				return nil, readErr
			}
			if sleepErr := sleepWithContext(ctx, backoff); sleepErr != nil {
				return nil, sleepErr
			}
			backoff *= 2
			continue
		}

		var out sendMultipleResponse
		decodeErr := json.Unmarshal(body, &out)
		if decodeErr != nil {
			lastErr = decodeErr
			if attempt == 4 {
				return nil, fmt.Errorf("send txs decode response: http=%d body=%s err=%w", resp.StatusCode, compactBody(body), decodeErr)
			}
			if sleepErr := sleepWithContext(ctx, backoff); sleepErr != nil {
				return nil, sleepErr
			}
			backoff *= 2
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusBadGateway || resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusGatewayTimeout {
			lastErr = fmt.Errorf("gateway send-multiple returned HTTP %d body=%s", resp.StatusCode, compactBody(body))
			if attempt == 4 {
				return nil, lastErr
			}
			if sleepErr := sleepWithContext(ctx, backoff); sleepErr != nil {
				return nil, sleepErr
			}
			backoff *= 2
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("gateway send-multiple returned HTTP %d error=%s code=%s body=%s", resp.StatusCode, out.Error, out.Code, compactBody(body))
		}

		if out.Error != "" && out.Error != "successful" {
			err := fmt.Errorf("gateway send-multiple error: %s (%s) http=%d body=%s", out.Error, out.Code, resp.StatusCode, compactBody(body))
			if attempt == 4 || !IsTransientSendError(err) {
				return nil, err
			}
			lastErr = err
			if sleepErr := sleepWithContext(ctx, backoff); sleepErr != nil {
				return nil, sleepErr
			}
			backoff *= 2
			continue
		}

		hashes := orderedTxHashes(out.Data.TxsHashes)
		if out.Data.NumOfSentTxs > 0 && out.Data.NumOfSentTxs != len(hashes) {
			return hashes, fmt.Errorf("gateway send-multiple accepted %d/%d transactions", out.Data.NumOfSentTxs, len(txs))
		}
		if len(hashes) == 0 {
			return nil, fmt.Errorf("gateway send-multiple returned no tx hashes")
		}
		return hashes, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("send txs failed after retries")
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func compactBody(body []byte) string {
	s := strings.TrimSpace(string(body))
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 400 {
		return s[:400] + "..."
	}
	return s
}

func orderedTxHashes(in map[string]string) []string {
	if len(in) == 0 {
		return nil
	}
	type item struct {
		idx  int
		hash string
	}
	items := make([]item, 0, len(in))
	for k, v := range in {
		idx := len(items)
		if parsed, err := strconv.Atoi(k); err == nil {
			idx = parsed
		}
		items = append(items, item{idx: idx, hash: v})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].idx == items[j].idx {
			return items[i].hash < items[j].hash
		}
		return items[i].idx < items[j].idx
	})
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item.hash != "" {
			out = append(out, item.hash)
		}
	}
	return out
}

type txStatusResponse struct {
	Data struct {
		Transaction struct {
			Status string `json:"status"`
		} `json:"transaction"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}

func (c *Client) GetTxStatus(ctx context.Context, txHash string) (string, error) {
	status, err := c.getTxStatusFromPath(ctx, fmt.Sprintf("%s/transactions/%s", c.baseURL, txHash))
	if err == nil {
		return NormalizeTxStatus(status), nil
	}

	status, err = c.getTxStatusFromPath(ctx, fmt.Sprintf("%s/transaction/%s", c.baseURL, txHash))
	if err != nil {
		return "", err
	}
	return NormalizeTxStatus(status), nil
}

func (c *Client) getTxStatusFromPath(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out txStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Error != "" && out.Error != "successful" {
		return "", fmt.Errorf("gateway error: %s (%s)", out.Error, out.Code)
	}
	if out.Data.Transaction.Status == "" {
		return "", fmt.Errorf("missing tx status in response from %s", url)
	}
	return out.Data.Transaction.Status, nil
}

func NormalizeTxStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

func IsFinalSuccessStatus(status string) bool {
	switch NormalizeTxStatus(status) {
	case "success", "successful", "executed":
		return true
	default:
		return false
	}
}

func IsFinalFailureStatus(status string) bool {
	switch NormalizeTxStatus(status) {
	case "fail", "failed", "invalid", "rejected":
		return true
	default:
		return false
	}
}

type accountSummaryResponse struct {
	Address string `json:"address"`
	Nonce   uint64 `json:"nonce"`
	Shard   int    `json:"shard"`
}

type accountToken struct {
	Identifier string `json:"identifier"`
	Balance    string `json:"balance"`
}

func (c *Client) GetAccountShard(ctx context.Context, bech32 string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/accounts/%s", c.baseURL, bech32), nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var out accountSummaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, err
	}
	if out.Address == "" {
		return 0, fmt.Errorf("missing account in response for %s", bech32)
	}
	return out.Shard, nil
}

func (c *Client) GetAccountTokenBalance(ctx context.Context, bech32 string, tokenID string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/accounts/%s/tokens", c.baseURL, bech32), nil)
	if err != nil {
		return "", err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out []accountToken
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	for _, token := range out {
		if token.Identifier == tokenID {
			return token.Balance, nil
		}
	}
	return "0", nil
}
