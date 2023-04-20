package util

import (
	"fmt"
	sdk "github.com/openshift-online/ocm-sdk-go"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
)

func GetSubscription(connection *sdk.Connection, key string) (*amv1.Subscription, error) {
	subsSearch := fmt.Sprintf(
		"(display_name = '%s' or cluster_id = '%s' or external_cluster_id = '%s' or id = '%s')",
		key, key, key, key)
	subsListResponse, err := connection.AccountsMgmt().V1().Subscriptions().List().Parameter("search", subsSearch).Send()
	if err != nil {
		err = fmt.Errorf("can't retrieve subscription for key '%s': %v", key, err)
		return nil, err
	}

	subsTotal := subsListResponse.Total()
	if subsTotal == 1 {
		return subsListResponse.Items().Get(0), nil
	}

	return nil, fmt.Errorf("there are %d subscriptions with cluster identifier or name '%s', expected 1", subsTotal, key)
}

func GetAccount(connection *sdk.Connection, key string) (*amv1.Account, error) {
	accsSearch := fmt.Sprintf("(username = '%s' or id = '%s')", key, key)
	accsListResponse, err := connection.AccountsMgmt().V1().Accounts().List().Parameter("search", accsSearch).Send()
	if err != nil {
		return nil, fmt.Errorf("can't retrieve account for key '%s': %v", key, err)
	}

	accsTotal := accsListResponse.Total()
	if accsTotal == 1 {
		return accsListResponse.Items().Get(0), nil
	}

	return nil, fmt.Errorf("there are %d accounts with id or username '%s', expected 1", accsTotal, key)
}
