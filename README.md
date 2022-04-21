# hedera-setup-go
Command line tool to perform basic setup on the hedera network.

## Running locally
Using zenv setup environment `source zenv hedera-sandbox`. Should have environment variables for the operator account id and private key.

### create account
```bash
go run cmd/main.go create-account \
    -network testnet \
    -operator-account $ACCOUNT_ID \
    -operator-private-key $PRIVATE_KEY
```

### create topic
```bash
go run cmd/main.go create-topic \
    -network testnet \
    -operator-account $ACCOUNT_ID \
    -operator-private-key $PRIVATE_KEY \
    -memo <memo_for_topic>
```

### account balance
```bash
go run cmd/main.go account-balance \
    -network testnet \
    -operator-account $ACCOUNT_ID \
    -operator-private-key $PRIVATE_KEY \
    -account-id <account_id_to_query>
```
