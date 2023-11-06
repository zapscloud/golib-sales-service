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
type MaterialTypeService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(materialTypeId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(materialTypeId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(materialTypeId string, delete_permanent bool) error

	EndService()
}

// BrandService - Brand Service structure
type materialTypeBaseService struct {
	db_utils.DatabaseService
	dbRegion        db_utils.DatabaseService
	daoMaterialType sales_repository.MaterialTypeDao
	daoBusiness     platform_repository.BusinessDao
	child           MaterialTypeService
	businessId      string
}

// NewMaterialTypeService - Construct MaterialType
func NewMaterialTypeService(props utils.Map) (MaterialTypeService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("MaterialTypeService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := materialTypeBaseService{}
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

// materialTypeBaseService - Close all the services
func (p *materialTypeBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *materialTypeBaseService) initializeService() {
	log.Printf("MaterialTypeService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoMaterialType = sales_repository.NewMaterialTypeDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *materialTypeBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("materialTypeBaseService::FindAll - Begin")

	listdata, err := p.daoMaterialType.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("materialTypeBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *materialTypeBaseService) Get(materialTypeId string) (utils.Map, error) {
	log.Printf("materialTypeBaseService::Get::  Begin %v", materialTypeId)

	data, err := p.daoMaterialType.Get(materialTypeId)

	log.Println("BrandService::Get:: End ", err)
	return data, err
}

func (p *materialTypeBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("materialTypeBaseService::FindByCode::  Begin ", filter)

	data, err := p.daoMaterialType.Find(filter)
	log.Println("materialTypeBaseService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *materialTypeBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("MaterialTypeService::Create - Begin")

	var materialTypeId string

	dataval, dataok := indata[sales_common.FLD_MATERIAL_TYPE_ID]
	if dataok {
		materialTypeId = strings.ToLower(dataval.(string))
	} else {
		materialTypeId = utils.GenerateUniqueId("mate")
		log.Println("Unique MaterialType ID", materialTypeId)
	}

	// Assign Business Id
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_MATERIAL_TYPE_ID] = materialTypeId

	data, err := p.daoMaterialType.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("MaterialTypeService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *materialTypeBaseService) Update(materialTypeId string, indata utils.Map) (utils.Map, error) {

	log.Println("MaterialTypeService::Update - Begin")

	data, err := p.daoMaterialType.Update(materialTypeId, indata)

	log.Println("MaterialTypeService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *materialTypeBaseService) Delete(materialTypeId string, delete_permanent bool) error {

	log.Println("BrandService::Delete - Begin", materialTypeId)

	if delete_permanent {
		result, err := p.daoMaterialType.Delete(materialTypeId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(materialTypeId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("BrandService::Delete - End")
	return nil
}

func (p *materialTypeBaseService) errorReturn(err error) (MaterialTypeService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
