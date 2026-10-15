# substrate-api-rpc

## Contents

- [Install](#Install)
- [Usage](#Usage)
    - [Codec](#Codec)
    - [RPC](#RPC)
- [Contributions](#Contributions)
- [LICENSE](#LICENSE)

## Install

```
go get github.com/itering/substrate-api-rpc
```

### Quip fork dependency pin

This fork retains the `github.com/itering/substrate-api-rpc` module identity.
Its decoder replacement retains the `github.com/itering/scale.go` import path and
resolves to `github.com/QuipNetwork/scale.go` at reviewed commit
`520629284934adca438e53c05fc1bce75a876e79`, represented by Go pseudo-version
`v0.0.0-20261014234944-520629284934`. The module checksums are committed in
`go.sum`; a local checkout or Docker decoder patch is not required.

The decoder requires Go 1.24 or newer, so this module declares Go 1.24.0.
Its dependency graph also selects `golang.org/x/crypto v0.45.0` and
`golang.org/x/sys v0.38.0`. These are inherited from the pinned decoder.

Go ignores dependency modules' `replace` directives. A consuming main module
must therefore carry this same decoder replacement as well as its RPC fork pin;
pinning only this RPC fork does not select the Quip decoder transitively.

Offline unit checks (after dependencies have been downloaded):

```sh
go test -mod=readonly . ./keyring ./util/...
go build -mod=readonly ./...
go vet -mod=readonly ./...
```

The `rpc` package's existing transaction tests initialize a public RPC connection
and submit transactions. Do not include them in the offline unit command above.

## Usage

### Codec

#### Extrinsic Decode

```
metadataRaw := "" // rpc state_getMetadata
specVersion := 0  // rpc chain_getRuntimeVersion
encodeExtrinsic := []string{"0x280402000b10449a7e7301", "0x1c0407005e8b4100"}
decodeExtrinsics, err := substrate.DecodeExtrinsic(encodeExtrinsic, metadata.Process(metadataRaw), specVersion)
```

#### Event Decode

```
metadataRaw := "" // rpc state_getMetadata
specVersion := 0  // rpc chain_getRuntimeVersion
event = "0x080000000000000080e36a09000000000200000001000000000000ca9a3b00000000020000"
substrate.DecodeEvent(event, metadataInstant, specVersion)
```

#### Log Decode

```
logs := ["0x054241424501014a7024ec6c4be378c35c254860d8f4ddc6f9d53ea8ce42ca00bc77c280511f1cb4c93fbd825e3c7dcabb36221372a9b5359c496e095d31afc359bdb9fac45487"]
substrate.DecodeLogDigest(logs)
```

#### Storage Decode

```
raw := "0x2efb"
storage.Decode(raw, "i16", nil)
```

### KeyRing

#### Generate KeyPair

```
keyRing := keyring.New(keyring.Sr25519Type, AliceSeed) // sr25519
keyRing := keyring.New(keyring.Ed25519Type, AliceSeed) // ed25519
```

#### Sign Message

```
keyRing.Sign("hello world") // sign utf-8 message
keyRing.Sign("0xffff")      // sign hex message
```

### RPC

#### Substrate RPC

Example

> state_getMetadata

```
blockHash := ""
rpc.GetMetadataByHash(conn, blockHash)
```

> state_getStorage

```
validatorsRaw, err := rpc.ReadStorage(conn, "Session", "Validators", blockHash)
validatorList := validatorsRaw.ToStringSlice()
```

#### Send Extrinsic

```
// set websocket endoint 
websocket.SetEndpoint("wss://shibuya-rpc.dwellir.com")

client = &rpc.Client{}

// init latest metadata
raw, err := GetMetadataByHash(nil)
if err != nil {
    panic(err)
}

// set metadata
client.SetMetadata(metadata.RegNewMetadataType(92, raw))

// set sr25519 seed
client.SetKeyRing(keyring.New(keyring.Sr25519Type, AliceSeed))

// sign transaction
signedTransaction, err := client.SignTransaction("Balances", "transfer", map[string]interface{}{"Id": BobAccountId}, 12345)

// send transaction async, will return block hash
blockHash, err := client.SendAuthorSubmitAndWatchExtrinsic(signedTransaction)

// send transaction synchronize, will return transaction hash
transactionHash, err := client.SendAuthorSubmitExtrinsic(signedTransaction)
```

More information can be viewed https://polkadot.js.org/api/substrate/rpc.html

## Contributions

We welcome contributions of any kind. Issues labeled can be good (first) contributions.

## LICENSE

Apache-2.0
