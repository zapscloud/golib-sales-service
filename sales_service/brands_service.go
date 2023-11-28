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

type BrandService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(brandId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(brandId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(brandId string, delete_permanent bool) error

	EndService()
}

type brandBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoBrand    sales_repository.BrandDao
	daoBusiness platform_repository.BusinessDao
	child       BrandService
	businessId  string
}

// NewBrandService - Construct Brand
func NewBrandService(props utils.Map) (BrandService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("BrandService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := brandBaseService{}
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

// brandsBaseService - Close all the service
func (p *brandBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *brandBaseService) initializeService() {
	log.Printf("BrandService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoBrand = sales_repository.NewBrandDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *brandBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("brandBaseService::FindAll - Begin")

	listdata, err := p.daoBrand.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("brandBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *brandBaseService) Get(brandId string) (utils.Map, error) {
	log.Printf("brandBaseService::Get::  Begin %v", brandId)

	data, err := p.daoBrand.Get(brandId)

	log.Println("brandBaseService::Get:: End ", err)
	return data, err
}

func (p *brandBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("brandService::FindByCode::  Begin ", filter)

	data, err := p.daoBrand.Find(filter)
	log.Println("brandService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *brandBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("BrandService::Create - Begin")
	var brandId string

	dataval, dataok := indata[sales_common.FLD_BRAND_ID]
	if dataok {
		brandId = strings.ToLower(dataval.(string))
	} else {
		brandId = utils.GenerateUniqueId("brnd")
		log.Println("Unique Brand ID", brandId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_BRAND_ID] = brandId

	data, err := p.daoBrand.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("BrandService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *brandBaseService) Update(brandId string, indata utils.Map) (utils.Map, error) {

	log.Println("BrandService::Update - Begin")

	data, err := p.daoBrand.Update(brandId, indata)

	log.Println("BrandService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *brandBaseService) Delete(brandId string, delete_permanent bool) error {

	log.Println("BrandService::Delete - Begin", brandId)

	if delete_permanent {
		result, err := p.daoBrand.Delete(brandId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(brandId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("BrandService::Delete - End")
	return nil
}

func (p *brandBaseService) errorReturn(err error) (BrandService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
