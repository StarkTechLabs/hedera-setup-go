package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/pkg/errors"
)

func main() {
	ctx := context.Background()

	createAccountCmd := flag.NewFlagSet("create-account", flag.ExitOnError)
	caNetwork := createAccountCmd.String("network", "testnet", "hedera network")
	caOpAcc := createAccountCmd.String("operator-account", "", "the operator account id")
	caOpKey := createAccountCmd.String("operator-private-key", "", "the operator private key")

	createTopicCmd := flag.NewFlagSet("create-topic", flag.ExitOnError)
	ctNetwork := createTopicCmd.String("network", "testnet", "hedera network")
	ctOpAcc := createTopicCmd.String("operator-account", "", "the operator account id")
	ctOpKey := createTopicCmd.String("operator-private-key", "", "the operator private key")
	ctMemo := createTopicCmd.String("memo", "test topic", "the memo of the topic to be created")

	if len(os.Args) < 2 {
		fmt.Println("expected subcommand")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "create-account":
		createAccountCmd.Parse(os.Args[2:])

		operatorAccount, err := hedera.AccountIDFromString(*caOpAcc)
		if err != nil {
			panic(err)
		}

		operatorKey, _ := hedera.PrivateKeyFromString(*caOpKey)
		if err != nil {
			panic(err)
		}

		client := setupClient(*caNetwork, operatorAccount, operatorKey)
		if client == nil {
			panic(errors.New("failed to create hedera client"))
		}

		if err := createAccount(ctx, client); err != nil {
			panic(err)
		}

		os.Exit(0)
	case "create-topic":
		createTopicCmd.Parse(os.Args[2:])

		operatorAccount, err := hedera.AccountIDFromString(*ctOpAcc)
		if err != nil {
			panic(err)
		}

		operatorKey, err := hedera.PrivateKeyFromString(*ctOpKey)
		if err != nil {
			panic(err)
		}

		client := setupClient(*ctNetwork, operatorAccount, operatorKey)
		if client == nil {
			panic(errors.New("failed to create hedera client"))
		}

		if err := createTopic(ctx, client, operatorKey, *ctMemo); err != nil {
			panic(err)
		}

		os.Exit(0)
	case "account-balance":
		cmd := flag.NewFlagSet("create-account", flag.ExitOnError)
		cmdNetwork := cmd.String("network", "testnet", "hedera network")
		cmdOperatorAccount := cmd.String("operator-account", "", "the operator account id")
		cmdOperatorPrivateKey := cmd.String("operator-private-key", "", "the operator private key")
		cmdAccountID := cmd.String("account-id", "", "hedera account id")

		cmd.Parse(os.Args[2:])

		if err := queryAccountBalance(ctx, cmdOperatorAccount, cmdOperatorPrivateKey, cmdAccountID, cmdNetwork); err != nil {
			panic(err)
		}

		os.Exit(0)
	default:
		fmt.Println("expected subcommand")
		os.Exit(1)
	}
}

func setupClient(network string, operatorAccount hedera.AccountID, operatorPrivateKey hedera.PrivateKey) *hedera.Client {
	var client *hedera.Client
	switch network {
	case "mainnet":
		client = hedera.ClientForMainnet()
	case "previewnet":
		client = hedera.ClientForPreviewnet()
	default:
		client = hedera.ClientForTestnet()
	}

	client.SetOperator(operatorAccount, operatorPrivateKey)
	return client
}

func createAccount(ctx context.Context, client *hedera.Client) error {
	fmt.Println("creating hedera account")

	txnId := hedera.TransactionIDGenerate(client.GetOperatorAccountID())

	privateKey, err := hedera.GeneratePrivateKey()
	if err != nil {
		return errors.Wrap(err, "failed to generate a new private key")
	}

	fmt.Println("building transaction")

	txn := hedera.NewAccountCreateTransaction().
		SetTransactionID(txnId).
		SetKey(privateKey.PublicKey()).
		SetInitialBalance(hedera.HbarFrom(0, hedera.HbarUnits.Tinybar))

	fmt.Println("executing transaction")

	txnID, err := txn.Execute(client)
	if err != nil {
		return errors.Wrap(err, "failed to execute transaction")
	}

	fmt.Println("getting receipt")

	receipt, err := txnID.GetReceipt(client)
	if err != nil {
		return errors.Wrap(err, "failed to get receipt of transaction")
	}

	if receipt.Status != hedera.StatusSuccess {
		return errors.New(fmt.Sprintf("Unable to create hedera account (receipt shows non-Success status %v)\n", receipt.Status))
	}

	accountID := receipt.AccountID

	fmt.Println("Account created.")
	fmt.Printf("Account ID: %s\n", accountID.String())
	fmt.Printf("Private Key: %s\n", privateKey.String())
	fmt.Printf("Public Key: %s\n", privateKey.PublicKey().String())
	fmt.Printf("Status: %s\n", receipt.Status.String())

	fmt.Printf("Details: %v\n", receipt)

	return nil
}

func createTopic(ctx context.Context, client *hedera.Client, operatorKey hedera.PrivateKey, memo string) error {
	fmt.Println("creating hedera topic")

	submitKey, err := hedera.GeneratePrivateKey()
	if err != nil {
		return errors.Wrap(err, "failed to generate a private topic submit key.")
	}

	// Build the Topic Create transaction, setting the keypairs we will use as well as some required values
	txn := hedera.NewTopicCreateTransaction().
		SetTopicMemo(memo).
		SetAdminKey(client.GetOperatorPublicKey()).
		SetSubmitKey(submitKey.PublicKey())

	fmt.Println("executing topic transaction")

	txnID, err := txn.Execute(client)
	if err != nil {
		return errors.Wrap(err, "failed to execute transaction")
	}

	receipt, err := txnID.GetReceipt(client)
	if err != nil {
		return errors.Wrap(err, "failed to get receipt of transaction")
	}

	if receipt.Status != hedera.StatusSuccess {
		return errors.New(fmt.Sprintf("Unable to create hedera topic (receipt shows non-Success status %v)\n", receipt.Status))
	}

	topicID := receipt.TopicID

	fmt.Println("Topic created.")
	fmt.Printf("Topic ID: %s\n", topicID.String())
	fmt.Println("--------------")
	fmt.Printf("Topic Submit Key: %s\n", submitKey.String())
	fmt.Println("--------------")
	fmt.Printf("Topic Running Hash: %s\n", string(receipt.TopicRunningHash))
	fmt.Printf("Topic Seq Number: %d\n", receipt.TopicSequenceNumber)

	return nil
}

func queryAccountBalance(ctx context.Context, cmdOperatorAccount, cmdOperatorPrivateKey, cmdAccountID, cmdNetwork *string) error {
	operatorAccount, err := hedera.AccountIDFromString(*cmdOperatorAccount)
	if err != nil {
		panic(err)
	}

	operatorKey, err := hedera.PrivateKeyFromString(*cmdOperatorPrivateKey)
	if err != nil {
		panic(err)
	}

	hAccountID, err := hedera.AccountIDFromString(*cmdAccountID)
	if err != nil {
		panic(err)
	}

	client := setupClient(*cmdNetwork, operatorAccount, operatorKey)
	if client == nil {
		panic(errors.New("failed to create hedera client"))
	}

	query := hedera.NewAccountBalanceQuery().SetAccountID(hAccountID)

	balance, err := query.Execute(client)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Account: %s\n", hAccountID.String())
	fmt.Printf("Balance: %s\n", balance.Hbars.String())

	return nil
}
