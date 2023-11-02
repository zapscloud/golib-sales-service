package sales_services

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

type DiscountService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(discountId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(discountId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(discountId string, delete_permanent bool) error

	EndService()
}

type discountBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoDiscount sales_repository.DiscountDao
	daoBusiness platform_repository.BusinessDao
	child       DiscountService
	businessId  string
}

// NewDiscountService - Construct Discount
func NewDiscountService(props utils.Map) (DiscountService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("DiscountService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := discountBaseService{}
	// Open Database Service
	err = p.OpenDatabaseService(props)
	if err != nil {
		return nil, err
	}

	// Open RegionDB Service
	p.dbRegion, err = platform_services.OpenRegionDatabaseService(props)
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

// discountsBaseService - Close all the services
func (p *discountBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *discountBaseService) initializeService() {
	log.Printf("DiscountService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoDiscount = sales_repository.NewDiscountDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *discountBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("discountBaseService::FindAll - Begin")

	listdata, err := p.daoDiscount.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("discountBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *discountBaseService) Get(discountId string) (utils.Map, error) {
	log.Printf("discountBaseService::Get::  Begin %v", discountId)

	data, err := p.daoDiscount.Get(discountId)

	log.Println("discountBaseService::Get:: End ", err)
	return data, err
}

func (p *discountBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("DiscountService::FindByCode::  Begin ", filter)

	data, err := p.daoDiscount.Find(filter)
	log.Println("DiscountService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *discountBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("DiscountService::Create - Begin")
	var discountId string

	dataval, dataok := indata[sales_common.FLD_DISCOUNT_ID]
	if dataok {
		discountId = strings.ToLower(dataval.(string))
	} else {
		discountId = utils.GenerateUniqueId("disc")
		log.Println("Unique Discount ID", discountId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_DISCOUNT_ID] = discountId

	data, err := p.daoDiscount.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("DiscountService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *discountBaseService) Update(discountId string, indata utils.Map) (utils.Map, error) {

	log.Println("DiscountService::Update - Begin")

	data, err := p.daoDiscount.Update(discountId, indata)

	log.Println("DiscountService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *discountBaseService) Delete(discountId string, delete_permanent bool) error {

	log.Println("DiscountService::Delete - Begin", discountId)

	if delete_permanent {
		result, err := p.daoDiscount.Delete(discountId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(discountId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("DiscountService::Delete - End")
	return nil
}

func (p *discountBaseService) errorReturn(err error) (DiscountService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
