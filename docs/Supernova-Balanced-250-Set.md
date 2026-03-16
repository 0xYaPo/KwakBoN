# Supernova Balanced 250 Wallet Set

Purpose:
- define a deterministic cross-shard active sender set for Supernova live testing
- compare a balanced cross-shard run against the previous shard-2-heavy runs

Source manifest:
- [wallets-manifest.json](/C:/Users/portyp/Mvx/KwakBoN/configs/wallets-manifest.json)

Derived manifest for direct use:
- [wallets-manifest.supernova-balanced-250.json](/C:/Users/portyp/Mvx/KwakBoN/configs/wallets-manifest.supernova-balanced-250.json)

## Target Composition

Balanced active set:
- shard `0`: `69`
- shard `1`: `90`
- shard `2`: `91`
- total: `250`

Selection policy:
- shard `0`: all available sender wallets
- shard `1`: first `90` sender wallets by `walletId`
- shard `2`: first `91` `active_candidate` wallets by `walletId`

Status used in the derived manifest:
- `supernova_balanced_active`

Extra tag added in the derived manifest:
- `supernova_balanced`

## Exact Wallet Selection

### Shard 0

All shard-0 sender wallets:
- `gw-254`
- `gw-266`
- `gw-268`
- `gw-271`
- `gw-272`
- `gw-279`
- `gw-280`
- `gw-292`
- `gw-294`
- `gw-297`
- `gw-299`
- `gw-301`
- `gw-305`
- `gw-307`
- `gw-308`
- `gw-312`
- `gw-314`
- `gw-316`
- `gw-317`
- `gw-322`
- `gw-323`
- `gw-328`
- `gw-329`
- `gw-338`
- `gw-339`
- `gw-341`
- `gw-346`
- `gw-358`
- `gw-365`
- `gw-366`
- `gw-368`
- `gw-370`
- `gw-374`
- `gw-377`
- `gw-378`
- `gw-381`
- `gw-383`
- `gw-390`
- `gw-391`
- `gw-402`
- `gw-403`
- `gw-404`
- `gw-406`
- `gw-412`
- `gw-417`
- `gw-423`
- `gw-424`
- `gw-428`
- `gw-431`
- `gw-432`
- `gw-436`
- `gw-441`
- `gw-442`
- `gw-445`
- `gw-446`
- `gw-450`
- `gw-454`
- `gw-456`
- `gw-460`
- `gw-462`
- `gw-470`
- `gw-473`
- `gw-479`
- `gw-483`
- `gw-488`
- `gw-491`
- `gw-494`
- `gw-496`
- `gw-497`

### Shard 1

First `90` shard-1 sender wallets by `walletId`:
- `gw-251`
- `gw-253`
- `gw-256`
- `gw-257`
- `gw-260`
- `gw-262`
- `gw-263`
- `gw-264`
- `gw-267`
- `gw-269`
- `gw-273`
- `gw-276`
- `gw-277`
- `gw-278`
- `gw-281`
- `gw-282`
- `gw-283`
- `gw-285`
- `gw-286`
- `gw-290`
- `gw-291`
- `gw-293`
- `gw-295`
- `gw-298`
- `gw-303`
- `gw-304`
- `gw-306`
- `gw-309`
- `gw-310`
- `gw-313`
- `gw-315`
- `gw-318`
- `gw-319`
- `gw-321`
- `gw-324`
- `gw-331`
- `gw-333`
- `gw-334`
- `gw-335`
- `gw-340`
- `gw-344`
- `gw-345`
- `gw-347`
- `gw-348`
- `gw-349`
- `gw-350`
- `gw-351`
- `gw-353`
- `gw-354`
- `gw-355`
- `gw-357`
- `gw-359`
- `gw-362`
- `gw-363`
- `gw-364`
- `gw-369`
- `gw-371`
- `gw-372`
- `gw-373`
- `gw-375`
- `gw-376`
- `gw-379`
- `gw-380`
- `gw-382`
- `gw-385`
- `gw-386`
- `gw-388`
- `gw-392`
- `gw-393`
- `gw-394`
- `gw-395`
- `gw-397`
- `gw-398`
- `gw-399`
- `gw-408`
- `gw-409`
- `gw-410`
- `gw-414`
- `gw-415`
- `gw-418`
- `gw-419`
- `gw-422`
- `gw-425`
- `gw-426`
- `gw-427`
- `gw-429`
- `gw-430`
- `gw-433`
- `gw-435`
- `gw-437`

### Shard 2

First `91` shard-2 `active_candidate` wallets by `walletId`:
- `gw-001` through `gw-091`

## Operational Use

The derived manifest marks the selected wallets with:
- `status=supernova_balanced_active`
- tag `supernova_balanced`

This allows direct testing without editing the primary manifest.

### Funding command shape

Use:
- `WALLETS_MANIFEST=./configs/wallets-manifest.supernova-balanced-250.json`
- `FUND_INCLUDE_STATUSES=supernova_balanced_active`
- `FUND_REQUIRED_TAGS=sender,supernova_balanced`
- `FUND_SHARD_FILTER=-1`

### Windowsprint command shape

Use:
- `WALLETS_MANIFEST=./configs/wallets-manifest.supernova-balanced-250.json`
- `SPRINT_SENDER_STATUSES=supernova_balanced_active`
- `SPRINT_REQUIRED_SENDER_TAGS=sender,supernova_balanced`
- `SPRINT_SHARD_FILTER=-1`

### Comparison target

Compare this set against:
- the previous shard-2-only `250` sender run

Suggested first A/B:
- same duration
- same TPS
- same sender engine
- only the wallet topology changes
