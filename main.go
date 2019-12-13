package payment

import (
	"encoding/json"
	"github.com/micro/go-micro/errors"
	"github.com/onskycloud/go-redis"
	"time"
)

// ServiceName Seperate Char
const ServiceName = "payment"

// SubscriptionKey Seperate Char
const SubscriptionKey = "onsky:payment:subscriptions"

const (
	Time     = iota + 1 // quota + time => duration>0, duration = endate - startdate
	Quota               // quota only => duration =0 => depend on service, nerver expired
	Interval            // quota + time , duration = (endate - startdate)/interval time
)

// Validate payment subscription
func Validate(db *redis.Redis, customerNumber string, serviceName string) (uint16,error) {
	now := time.Now()
	if db == nil || customerNumber == "" || serviceName == ""{
		return 0,errors.BadRequest(ServiceName, "validate:invalidParams")
	}
	subscription, err := getSubscription(db, customerNumber, serviceName)
	if err != nil {
		return 0,err
	}
	duration,err := checkSubscription(subscription, now)
	if err != nil{
		return 0,err
	}
	if duration == 0{
		return 0,errors.Forbidden(ServiceName, "validate:expired")
	}
	return duration,nil
}
func checkSubscription(subscription *SubscriptionCache, now time.Time) (uint16,error) {
	if subscription.Status != 2 { //hardcode
		return 0,errors.Forbidden(ServiceName, "validate:inactiveSubscription")
	}
	if subscription.OldPrice <= 0 {
		return 0,errors.Forbidden(ServiceName, "validate:invalidPrice")
	}
	if subscription.Quota <= 0 {
		return 0,errors.Forbidden(ServiceName, "validate:invalidQuota")
	}
	switch subscription.Type {
	case Quota:
		if subscription.Duration != 0 {
			return 0,errors.Forbidden(ServiceName, "validate:invalidDuration")
		}
		if subscription.IntervalTime != 0 {
			return 0,errors.Forbidden(ServiceName, "validate:invalidIntervalTime")
		}
		quota := subscription.Quota
		if subscription.HaveTrialPackage{
			if subscription.TrialDuration == 0{
				return 0,errors.Forbidden(ServiceName, "validate:invalidTrialDuration")
			}
			quota += subscription.TrialDuration
		}
		return quota,nil
	case Time:
		if subscription.StartDate.After(now) || subscription.EndDate.Before(now) || subscription.EndDate.Before(subscription.StartDate) {
			return 0,errors.Forbidden(ServiceName, "validate:invalidDate")
		}
		if subscription.IntervalTime != 0 {
			return 0,errors.Forbidden(ServiceName, "validate:invalidIntervalTime")
		}
		if subscription.Duration <= 0 {
			return 0,errors.Forbidden(ServiceName, "validate:invalidDuration")
		}
		duration := subscription.Duration
		if subscription.HaveTrialPackage{
			if subscription.TrialDuration == 0{
				return 0,errors.Forbidden(ServiceName, "validate:invalidTrialDuration")
			}
			duration += subscription.TrialDuration
		}
		if subscription.Duration > uint16(subscription.EndDate.Sub(subscription.StartDate).Hours()) {
			return 0,errors.Forbidden(ServiceName, "validate:invalidDuration")
		}
		return subscription.Quota,nil
	case Interval:
		if subscription.StartDate.After(now) || subscription.EndDate.Before(now) || subscription.EndDate.Before(subscription.StartDate) {
			return 0,errors.Forbidden(ServiceName, "validate:invalidDate")
		}
		if subscription.IntervalTime <= 0 {
			return 0,errors.Forbidden(ServiceName, "validate:invalidIntervalTime")
		}
		rangeDate:= int(subscription.EndDate.Sub(subscription.StartDate).Hours())
		day:=now.Day()
		invervalHour := int(subscription.IntervalTime) * day*24
		duration := rangeDate / invervalHour
		if subscription.Duration <= 0 || int(subscription.Duration) != duration {
			return 0,errors.Forbidden(ServiceName, "validate:invalidDuration")
		}
		return subscription.Quota,nil
	default:
		return 0,errors.Forbidden(ServiceName, "validate:invalidType")
	}
	return 0,errors.Forbidden(ServiceName, "validate:invalidType")
}
func getSubscription(db *redis.Redis, customerNumber string, serviceName string) (*SubscriptionCache, error) {
	var subscription = new(SubscriptionCache)
	var t interface{}
	// get subscription from cache
	err := db.GetObject(SubscriptionKey, customerNumber+serviceName, subscription)
	if err != nil {
		if err.Error()=="redis: nil"{
			return nil,errors.Forbidden(ServiceName, "validate:unregestered")
		}
		return nil, err
	}
	buffer, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(buffer, subscription)
	if subscription == nil {
		return nil, errors.InternalServerError(ServiceName, "validate:wrongData")
	}
	if subscription.CustomerNumber != customerNumber || subscription.Service != serviceName{
		return nil, errors.BadRequest(ServiceName, "validate:wrongCustomer")
	}
	return subscription, nil
}

//SubscriptionCache cache field customer_number+service_name
type SubscriptionCache struct {
	UUID             string                 `json:"uuid"`
	PackageID        string                 `json:"package_id"`
	StartDate        time.Time              `json:"start_date"`
	EndDate          time.Time              `json:"end_date"`
	Type             uint8                  `json:"type"`
	Meta             map[string]interface{} `json:"meta"`
	Duration         uint16                 `json:"duration"` // duration = hour
	IntervalTime     uint16                 `json:"interval_time"` // month
	Quota            uint16                 `json:"quota"`
	OldPrice         float64                `json:"old_price"`
	CustomerNumber   string                 `json:"customer_number"`
	Service          string                 `json:"service"`
	Status           uint8                  `json:"status"`
	HaveTrialPackage bool                   `json:"have_trial_package"`
	TrialDuration    uint16                 `json:"trial_duration"` // hour or quota
}
