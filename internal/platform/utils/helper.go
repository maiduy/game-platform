package util

import (
	"fmt"
	"time"
)

// GetPartitionedCollectionName generates a partitioned collection name based on the timestamp
// Format: <base_collection_name>_<yyyymm>
// Example: GetPartitionedCollectionName("profile_credit_scoring", 1700000000) -> "profile_credit_scoring_202311"
func GetPartitionedCollectionName(baseCollectionName string, timestamp int64) string {
	t := time.Unix(timestamp, 0)
	partition := t.Format("200601") // Format: yyyymm
	return fmt.Sprintf("%s_%s", baseCollectionName, partition)
}
