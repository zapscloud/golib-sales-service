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
type CatalogueService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(catalogueId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(catalogueId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(catalogueId string, delete_permanent bool) error

	EndService()
}

// BrandService - Brand Service structure
type catalogueBaseService struct {
	db_utils.DatabaseService
	dbRegion     db_utils.DatabaseService
	daoCatalogue sales_repository.CatalogueDao
	daoBusiness  platform_repository.BusinessDao
	child        CatalogueService
	businessId   string
}

// NewCatalogueService - Construct Catalogue
func NewCatalogueService(props utils.Map) (CatalogueService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("CatalogueService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := catalogueBaseService{}
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

// catalogueBaseService - Close all the service
func (p *catalogueBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *catalogueBaseService) initializeService() {
	log.Printf("CatalogueService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoCatalogue = sales_repository.NewCatalogueDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *catalogueBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("catalogueBaseService::FindAll - Begin")

	listdata, err := p.daoCatalogue.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("catalogueBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *catalogueBaseService) Get(catalogueId string) (utils.Map, error) {
	log.Printf("catalogueBaseService::Get::  Begin %v", catalogueId)

	data, err := p.daoCatalogue.Get(catalogueId)

	log.Println("BrandService::Get:: End ", err)
	return data, err
}

func (p *catalogueBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("catalogueBaseService::FindByCode::  Begin ", filter)

	data, err := p.daoCatalogue.Find(filter)
	log.Println("catalogueBaseService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *catalogueBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("CatalogueService::Create - Begin")

	var catalogueId string

	dataval, dataok := indata[sales_common.FLD_CATALOGUE_ID]
	if dataok {
		catalogueId = strings.ToLower(dataval.(string))
	} else {
		catalogueId = utils.GenerateUniqueId("catlg")
		log.Println("Unique Catalogue ID", catalogueId)
	}

	// Assign Business Id
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_CATALOGUE_ID] = catalogueId

	data, err := p.daoCatalogue.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("CatalogueService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *catalogueBaseService) Update(catalogueId string, indata utils.Map) (utils.Map, error) {

	log.Println("CatalogueService::Update - Begin")

	data, err := p.daoCatalogue.Update(catalogueId, indata)

	log.Println("CatalogueService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *catalogueBaseService) Delete(catalogueId string, delete_permanent bool) error {

	log.Println("BrandService::Delete - Begin", catalogueId)

	if delete_permanent {
		result, err := p.daoCatalogue.Delete(catalogueId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(catalogueId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("BrandService::Delete - End")
	return nil
}

func (p *catalogueBaseService) errorReturn(err error) (CatalogueService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
