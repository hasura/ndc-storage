package gcs

import (
	"cloud.google.com/go/storage"
	"github.com/hasura/ndc-storage/connector/storage/common"
)

func validateLifecycleRule(rule common.BucketLifecycleRule) storage.LifecycleRule {
	r := storage.LifecycleRule{
		Action: storage.LifecycleAction{},
		// ID:         rule.ID,
		// Expiration: validateLifecycleExpiration(rule.Expiration),
		// RuleFilter: validateLifecycleFilter(rule.RuleFilter),
		// Transition: validateLifecycleTransition(rule.Transition),
	}

	if rule.Expiration != nil && rule.Expiration.Days != nil {
		r.Condition.AgeInDays = int64(*rule.Expiration.Days)
		r.Condition.AllObjects = *rule.Expiration.Days == 0
	}

	if rule.Transition != nil || rule.Transition.StorageClass != nil {
		r.Action.StorageClass = *rule.Transition.StorageClass
	}

	// if rule.AbortIncompleteMultipartUpload != nil && rule.AbortIncompleteMultipartUpload.DaysAfterInitiation != nil {
	// 	r.AbortIncompleteMultipartUpload.DaysAfterInitiation = lifecycle.ExpirationDays(*rule.AbortIncompleteMultipartUpload.DaysAfterInitiation)
	// }

	// if rule.AllVersionsExpiration != nil && (rule.AllVersionsExpiration.Days != nil || rule.AllVersionsExpiration.DeleteMarker != nil) {
	// 	if rule.AllVersionsExpiration.Days != nil {
	// 		r.AllVersionsExpiration.Days = *rule.AllVersionsExpiration.Days
	// 	}

	// 	if rule.DelMarkerExpiration != nil {
	// 		r.AllVersionsExpiration.DeleteMarker = lifecycle.ExpireDeleteMarker(*rule.AllVersionsExpiration.DeleteMarker)
	// 	}
	// }

	// if rule.DelMarkerExpiration != nil && rule.DelMarkerExpiration.Days != nil {
	// 	r.DelMarkerExpiration.Days = *rule.DelMarkerExpiration.Days
	// }

	// if rule.NoncurrentVersionExpiration != nil {
	// 	if rule.NoncurrentVersionExpiration.NewerNoncurrentVersions != nil {
	// 		r.NoncurrentVersionExpiration.NewerNoncurrentVersions = *rule.NoncurrentVersionExpiration.NewerNoncurrentVersions
	// 	}

	// 	if rule.NoncurrentVersionExpiration.NoncurrentDays != nil {
	// 		r.NoncurrentVersionExpiration.NoncurrentDays = lifecycle.ExpirationDays(*rule.NoncurrentVersionExpiration.NoncurrentDays)
	// 	}
	// }

	// if rule.NoncurrentVersionTransition != nil {
	// 	if rule.NoncurrentVersionTransition.NewerNoncurrentVersions != nil {
	// 		r.NoncurrentVersionTransition.NewerNoncurrentVersions = *rule.NoncurrentVersionTransition.NewerNoncurrentVersions
	// 	}

	// 	if rule.NoncurrentVersionTransition.NoncurrentDays != nil {
	// 		r.NoncurrentVersionTransition.NoncurrentDays = lifecycle.ExpirationDays(*rule.NoncurrentVersionTransition.NoncurrentDays)
	// 	}

	// 	if rule.NoncurrentVersionTransition.StorageClass != nil {
	// 		r.NoncurrentVersionTransition.StorageClass = *rule.NoncurrentVersionTransition.StorageClass
	// 	}
	// }

	// if rule.Prefix != nil {
	// 	r.Prefix = *rule.Prefix
	// }

	// if rule.Status != nil {
	// 	r.Status = *rule.Status
	// }

	return r
}

func validateLifecycleConfiguration(input common.BucketLifecycleConfiguration) *storage.Lifecycle {
	result := &storage.Lifecycle{
		Rules: make([]storage.LifecycleRule, len(input.Rules)),
	}

	for i, rule := range input.Rules {
		r := validateLifecycleRule(rule)
		result.Rules[i] = r
	}

	return result
}

func serializeLifecycleConfiguration(input storage.Lifecycle) *common.BucketLifecycleConfiguration {
	result := &common.BucketLifecycleConfiguration{
		Rules: make([]common.BucketLifecycleRule, len(input.Rules)),
	}

	// for i, rule := range input.Rules {
	// 	// r := serializeLifecycleRule(rule)
	// 	result.Rules[i] = r
	// }

	return result
}
