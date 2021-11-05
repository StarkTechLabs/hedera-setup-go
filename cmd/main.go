package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/hashgraph/hedera-sdk-go"
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

		operatorAccount, _ := hedera.AccountIDFromString(*caOpAcc)
		operatorKey, _ := hedera.Ed25519PrivateKeyFromString(*caOpKey)

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

		operatorAccount, _ := hedera.AccountIDFromString(*ctOpAcc)
		operatorKey, _ := hedera.Ed25519PrivateKeyFromString(*ctOpKey)

		client := setupClient(*ctNetwork, operatorAccount, operatorKey)
		if client == nil {
			panic(errors.New("failed to create hedera client"))
		}

		if err := createTopic(ctx, client, operatorKey, *ctMemo); err != nil {
			panic(err)
		}

		os.Exit(0)
	default:
		fmt.Println("expected subcommand")
		os.Exit(1)
	}

}

func setupClient(network string, operatorAccount hedera.AccountID, operatorPrivateKey hedera.Ed25519PrivateKey) *hedera.Client {
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
	txnId := hedera.NewTransactionID(client.GetOperatorID())

	privateKey, err := hedera.GenerateEd25519PrivateKey()
	if err != nil {
		return errors.Wrap(err, "failed to generate a new private key")
	}

	txn, err := hedera.NewAccountCreateTransaction().
		SetTransactionID(txnId).
		SetKey(privateKey.PublicKey()).
		SetInitialBalance(hedera.HbarFrom(0, hedera.HbarUnits.Tinybar)).
		Build(client)
	if err != nil {
		return errors.Wrap(err, "failed to build create account transaction")
	}

	txnID, err := txn.Execute(client)
	if err != nil {
		return errors.Wrap(err, "failed to execute transaction")
	}

	receipt, err := txnID.GetReceipt(client)
	if err != nil {
		return errors.Wrap(err, "failed to get receipt of transaction")
	}

	if receipt.Status != hedera.StatusSuccess {
		return errors.New(fmt.Sprintf("Unable to create hedera account (receipt shows non-Success status %v)\n", receipt.Status))
	}

	accountID := receipt.GetAccountID()

	fmt.Println("Account created.")
	fmt.Printf("Account ID: %s\n", accountID.String())
	fmt.Printf("Private Key: %s\n", privateKey.String())
	fmt.Printf("Public Key: %s\n", privateKey.PublicKey().String())
	fmt.Printf("Status: %s\n", receipt.Status.String())

	fmt.Printf("Details: %v\n", receipt)

	return nil
}

func createTopic(ctx context.Context, client *hedera.Client, operatorKey hedera.Ed25519PrivateKey, memo string) error {
	adminKey, err := hedera.GenerateEd25519PrivateKey()
	if err != nil {
		return errors.Wrap(err, "failed to generate a private topic admin key.")
	}

	submitKey, err := hedera.GenerateEd25519PrivateKey()
	if err != nil {
		return errors.Wrap(err, "failed to generate a private topic submit key.")
	}

	// Build the Topic Create transaction, setting the keypairs we will use as well as some required values
	txn, err := hedera.NewConsensusTopicCreateTransaction().
		SetMaxTransactionFee(hedera.HbarFromTinybar(100000000)).
		SetTopicMemo(memo).
		SetAdminKey(adminKey.PublicKey()).
		SetSubmitKey(submitKey.PublicKey()).
		SetAutoRenewAccountID(client.GetOperatorID()).
		Build(client)
	if err != nil {
		return errors.Wrap(err, "failed to build topic create transaction")
	}

	txnID, err := txn.
		SignWith(operatorKey.PublicKey(), operatorKey.Sign).
		SignWith(adminKey.PublicKey(), adminKey.Sign).
		Execute(client)
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

	topicID := receipt.GetConsensusTopicID()
	// data := map[string]string{
	// 	"topicID": topicID.String(),
	// 	"topicSubmitKey": submitKey.String(),
	// 	"topicAdminKey": adminKey.String(),
	// }

	fmt.Println("Topic created.")
	fmt.Printf("Topic ID: %s\n", topicID.String())
	fmt.Println("--------------")
	fmt.Printf("Topic Submit Key: %s\n", submitKey.String())
	fmt.Println("--------------")
	fmt.Printf("Topic Admin Key: %s\n", adminKey.String())
	fmt.Println("--------------")

	return nil
}
