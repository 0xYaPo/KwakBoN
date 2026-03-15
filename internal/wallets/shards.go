package wallets

const NumberOfShardsWithoutMeta = 3

func ShardOfPubKey(pubKey []byte) int {
	if len(pubKey) == 0 {
		return -1
	}
	maskHigh := byte(0b11)
	maskLow := byte(0b01)
	lastByte := pubKey[len(pubKey)-1]
	shard := int(lastByte & maskHigh)
	if shard > NumberOfShardsWithoutMeta-1 {
		shard = int(lastByte & maskLow)
	}
	return shard
}
