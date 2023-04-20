package util

import (
	"fmt"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"regexp"
)

func IsValidClusterKey(clusterKey string) bool {
	return regexp.MustCompile(`^(\w|-)+$`).MatchString(clusterKey)
}

func GetCluster(connection *sdk.Connection, clusterId string) (*cmv1.Cluster, error) {
	clustersSearch := fmt.Sprintf("id = '%s' or name = '%s' or external_id = '%s'", clusterId, clusterId, clusterId)
	clustersListResponse, err := connection.ClustersMgmt().V1().Clusters().List().Search(clustersSearch).Size(1).Send()
	if err != nil {
		return nil, fmt.Errorf("can't retrieve clusters for clusterId '%s': %v", clusterId, err)
	}

	clustersTotal := clustersListResponse.Total()
	if clustersTotal == 1 {
		return clustersListResponse.Items().Slice()[0], nil
	}

	return nil, fmt.Errorf("there are %d clusters with identifier or name '%s', expected 1", clustersTotal, clusterId)
}

func GetActiveCluster(connection *sdk.Connection, key string) (*cmv1.Cluster, error) {
	subsSearch := fmt.Sprintf(
		"(display_name = '%s' or cluster_id = '%s' or external_cluster_id = '%s') and "+
			"status in ('Reserved', 'Active')",
		key, key, key,
	)
	subsListResponse, err := connection.AccountsMgmt().V1().Subscriptions().List().Search(subsSearch).Size(1).Send()
	if err != nil {
		return nil, fmt.Errorf("can't retrieve subscription for key '%s': %v", key, err)
	}

	subsTotal := subsListResponse.Total()
	if subsTotal == 1 {
		id, ok := subsListResponse.Items().Slice()[0].GetClusterID()
		if ok {
			var clusterGetResponse *cmv1.ClusterGetResponse
			clusterGetResponse, err = connection.ClustersMgmt().V1().Clusters().Cluster(id).Get().Send()
			if err != nil {
				return nil, fmt.Errorf("can't retrieve cluster for key '%s': %v", key, err)
			}
			return clusterGetResponse.Body(), nil
		}
	}

	if subsTotal > 1 {
		return nil, fmt.Errorf("there are %d subscriptions with cluster identifier or name '%s'", subsTotal, key)
	}

	// If we are here then no subscription matches the passed key. It may still be possible that
	// the cluster exists but is not reporting metrics, so it will not have the external
	// identifier in the accounts management service. To find those clusters we need to check
	// directly in the clusters management service.
	clustersSearch := fmt.Sprintf(
		"id = '%s' or name = '%s' or external_id = '%s'",
		key, key, key,
	)
	clustersListResponse, err := connection.ClustersMgmt().V1().Clusters().List().Search(clustersSearch).Size(1).Send()
	if err != nil {
		return nil, fmt.Errorf("can't retrieve clusters for key '%s': %v", key, err)
	}

	clustersTotal := clustersListResponse.Total()
	if clustersTotal == 1 {
		return clustersListResponse.Items().Slice()[0], nil
	}

	return nil, fmt.Errorf("there are %d clusters with identifier or name '%s', expected 1", clustersTotal, key)
}

func IsClusterCCS(connection *sdk.Connection, clusterID string) (bool, error) {
	clusterResponse, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).Get().Send()
	if err != nil {
		return false, err
	}

	cluster := clusterResponse.Body()
	if cluster.CCS().Enabled() {
		return true, nil
	}
	return false, nil
}

func GetHiveShard(connection *sdk.Connection, clusterID string) (string, error) {
	shardPath, err := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID).ProvisionShard().Get().Send()

	if err != nil {
		return "", err
	}

	var shard string

	if shardPath != nil {
		shard = shardPath.Body().HiveConfig().Server()
	}

	if shard == "" {
		return "", fmt.Errorf("Unable to retrieve shard for cluster %s", clusterID)
	}

	return shard, nil
}
