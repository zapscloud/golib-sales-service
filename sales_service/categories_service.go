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

// BrandService - Brand Service structure
type CategoryService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(categoryId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(categoryId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(categoryId string, delete_permanent bool) error

	EndService()
}

// BrandService - Brand Service structure
type categoryBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoCategory sales_repository.CategoryDao
	daoBusiness platform_repository.BusinessDao
	child       CategoryService
	businessId  string
}

// NewCategoryService - Construct Category
func NewCategoryService(props utils.Map) (CategoryService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("CategoryService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := categoryBaseService{}
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

// catogoryBaseService - Close all the service
func (p *categoryBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *categoryBaseService) initializeService() {
	log.Printf("CategoryService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoCategory = sales_repository.NewCategoryDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *categoryBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("categoryBaseService::FindAll - Begin")

	listdata, err := p.daoCategory.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("categoryBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *categoryBaseService) Get(categoryId string) (utils.Map, error) {
	log.Printf("categoryBaseService::Get::  Begin %v", categoryId)

	data, err := p.daoCategory.Get(categoryId)

	log.Println("BrandService::Get:: End ", err)
	return data, err
}

func (p *categoryBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("categoryBaseService::FindByCode::  Begin ", filter)

	data, err := p.daoCategory.Find(filter)
	log.Println("categoryBaseService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *categoryBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("CategoryService::Create - Begin")

	var categoryId string

	dataval, dataok := indata[sales_common.FLD_CATEGORY_ID]
	if dataok {
		categoryId = strings.ToLower(dataval.(string))
	} else {
		categoryId = utils.GenerateUniqueId("catg")
		log.Println("Unique Category ID", categoryId)
	}

	// Assign Business Id
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_CATEGORY_ID] = categoryId

	data, err := p.daoCategory.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("CategoryService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *categoryBaseService) Update(categoryId string, indata utils.Map) (utils.Map, error) {

	log.Println("CategoryService::Update - Begin")

	data, err := p.daoCategory.Update(categoryId, indata)

	log.Println("CategoryService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *categoryBaseService) Delete(categoryId string, delete_permanent bool) error {

	log.Println("BrandService::Delete - Begin", categoryId)

	if delete_permanent {
		result, err := p.daoCategory.Delete(categoryId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(categoryId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("BrandService::Delete - End")
	return nil
}

func (p *categoryBaseService) errorReturn(err error) (CategoryService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
