package sales_service

import (
	"fmt"
	"log"
	"strings"

	"github.com/zapscloud/golib-dbutils/db_common"
	"github.com/zapscloud/golib-dbutils/db_utils"
	"github.com/zapscloud/golib-platform-repository/platform_repository"
	"github.com/zapscloud/golib-platform-service/platform_service"
	"github.com/zapscloud/golib-sales-repository/sales_common"
	"github.com/zapscloud/golib-sales-repository/sales_repository"
	"github.com/zapscloud/golib-utils/utils"
)

type CouponService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(couponId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(couponId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(couponId string, delete_permanent bool) error

	EndService()
}

type couponBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoCoupon   sales_repository.CouponDao
	daoBusiness platform_repository.BusinessDao
	child       CouponService
	businessId  string
}

// NewCouponService - Construct Coupon
func NewCouponService(props utils.Map) (CouponService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("CouponService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := couponBaseService{}
	// Open Database Service
	err = p.OpenDatabaseService(props)
	if err != nil {
		return nil, err
	}

	// Open RegionDB Service
	p.dbRegion, err = platform_service.OpenRegionDatabaseService(props)
	if err != nil {
		p.CloseDatabaseService()
		return nil, err
	}

	// Assign the BusinessId
	p.businessId = businessId
	p.initializeService()

	_, err = p.daoBusiness.Get(businessId)
	if err != nil {
		err := &utils.AppError{
			ErrorCode:   funcode + "01",
			ErrorMsg:    "Invalid BusinessId",
			ErrorDetail: "Given BusinessId is not exist"}
		return p.errorReturn(err)
	}

	p.child = &p

	return &p, err
}

// couponBaseService - Close all the services
func (p *couponBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *couponBaseService) initializeService() {
	log.Printf("CouponService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoCoupon = sales_repository.NewCouponDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *couponBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("CouponBaseService::FindAll - Begin")

	listdata, err := p.daoCoupon.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("CouponBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *couponBaseService) Get(couponId string) (utils.Map, error) {
	log.Printf("couponBaseService::Get::  Begin %v", couponId)

	data, err := p.daoCoupon.Get(couponId)

	log.Println("couponBaseService::Get:: End ", err)
	return data, err
}

func (p *couponBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("couponBaseService::FindByCode::  Begin ", filter)

	data, err := p.daoCoupon.Find(filter)
	log.Println("couponBaseService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *couponBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("CouponService::Create - Begin")
	var couponId string

	dataval, dataok := indata[sales_common.FLD_COUPON_ID]
	if dataok {
		couponId = strings.ToLower(dataval.(string))
	} else {
		couponId = utils.GenerateUniqueId("coup")
		log.Println("Unique Coupon ID", couponId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_COUPON_ID] = couponId

	data, err := p.daoCoupon.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("CouponService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *couponBaseService) Update(couponId string, indata utils.Map) (utils.Map, error) {

	log.Println("CouponService::Update - Begin")

	data, err := p.daoCoupon.Update(couponId, indata)

	log.Println("CouponService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *couponBaseService) Delete(couponId string, delete_permanent bool) error {

	log.Println("CouponService::Delete - Begin", couponId)

	if delete_permanent {
		result, err := p.daoCoupon.Delete(couponId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(couponId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("CouponService::Delete - End")
	return nil
}

func (p *couponBaseService) errorReturn(err error) (CouponService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
