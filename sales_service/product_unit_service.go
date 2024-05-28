package sales_service

import (
	"fmt"
	"log"

	"github.com/zapscloud/golib-dbutils/db_common"
	"github.com/zapscloud/golib-dbutils/db_utils"
	"github.com/zapscloud/golib-platform-repository/platform_repository"
	"github.com/zapscloud/golib-platform-service/platform_service"
	"github.com/zapscloud/golib-utils/utils"
	"github.com/zapscloud/golib-sales-repository/sales_common"
	"github.com/zapscloud/golib-sales-repository/sales_repository"
)

const (
// // The character encoding for the email.
// CharSet = "UTF-8"
)

// Product_unitService - Product_unit Service structure
type Product_unitService interface {
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	Get(product_unit_id string) (utils.Map, error)
	Find(filter string) (utils.Map, error)
	Create(indata utils.Map) (utils.Map, error)
	Update(product_unit_id string, indata utils.Map) (utils.Map, error)
	Delete(product_unit_id string, delete_permanent bool) error

	BeginTransaction()
	CommitTransaction()
	RollbackTransaction()

	EndService()
}

// LoyaltyCardService - Product_unit Service structure
type Product_unitBaseService struct {
	dbRegion db_utils.DatabaseService
	db_utils.DatabaseService
	daoProduct_unit sales_repository.Product_unitDao
	daoBusiness     platform_repository.BusinessDao
	child           Product_unitService
	businessID      string
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
}

func NewProduct_unitService(props utils.Map) (Product_unitService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	p := Product_unitBaseService{}

	err := p.OpenDatabaseService(props)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Product_unitMongoService ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return p.errorReturn(err)
	}

	// Open RegionDB Service
	p.dbRegion, err = platform_service.OpenRegionDatabaseService(props)
	if err != nil {
		p.CloseDatabaseService()
		return nil, err
	}
	// Assign the BusinessId
	p.businessID = businessId
	p.initializeService()

	_, err = p.daoBusiness.Get(businessId)
	if err != nil {
		err := &utils.AppError{ErrorCode: funcode + "01", ErrorMsg: "Invalid business_id", ErrorDetail: "Given business_id is not exist"}
		return p.errorReturn(err)
	}

	p.child = &p

	return &p, err
}

// EndLoyaltyCardService - Close all the services
func (p *Product_unitBaseService) EndService() {
	log.Printf("EndProduct_unitService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *Product_unitBaseService) initializeService() {
	log.Printf("Product_unitService:: GetBusinessDao ")
	p.daoProduct_unit = sales_repository.NewProduct_unitDao(p.dbRegion.GetClient(), p.businessID)
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
}

// List - List All records
func (p *Product_unitBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("Product_unitService::FindAll - Begin")

	daoProduct_unit := p.daoProduct_unit
	response, err := daoProduct_unit.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("Product_unitService::FindAll - End ")
	return response, nil
}

// FindByCode - Find By Code
func (p *Product_unitBaseService) Get(product_unit_id string) (utils.Map, error) {
	log.Printf("Product_unitService::FindByCode::  Begin %v", product_unit_id)

	data, err := p.daoProduct_unit.Get(product_unit_id)
	log.Println("Product_unitService::FindByCode:: End ", err)
	return data, err
}

func (p *Product_unitBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("Product_unitService::FindByCode::  Begin ", filter)

	data, err := p.daoProduct_unit.Find(filter)
	log.Println("Product_unitService::FindByCode:: End ", data, err)
	return data, err
}

func (p *Product_unitBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("UserService::Create - Begin")

	dataval, dataok := indata[sales_common.FLD_PRODUCT_UNIT_ID]
	if !dataok {
		uid := utils.GenerateUniqueId("prdut")
		log.Println("Unique Product_unit ID", uid)
		indata[sales_common.FLD_PRODUCT_UNIT_ID] = uid
		dataval = indata[sales_common.FLD_PRODUCT_UNIT_ID]
	}
	indata[sales_common.FLD_BUSINESS_ID] = p.businessID
	log.Println("Provided Product_unit ID:", dataval)

	_, err := p.daoProduct_unit.Get(dataval.(string))
	if err == nil {
		err := &utils.AppError{ErrorCode: "S30102", ErrorMsg: "Existing Product_unit ID !", ErrorDetail: "Given Product_unit ID already exist"}
		return indata, err
	}

	insertResult, err := p.daoProduct_unit.Create(indata)
	if err != nil {
		return indata, err
	}
	log.Println("UserService::Create - End ", insertResult)
	return indata, err
}

// Update - Update Service
func (p *Product_unitBaseService) Update(product_unit_id string, indata utils.Map) (utils.Map, error) {

	log.Println("Product_unitService::Update - Begin")

	data, err := p.daoProduct_unit.Get(product_unit_id)
	if err != nil {
		return data, err
	}

	data, err = p.daoProduct_unit.Update(product_unit_id, indata)
	log.Println("Product_unitService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *Product_unitBaseService) Delete(product_unit_id string, delete_permanent bool) error {

	log.Println("Product_unit Service::Delete - Begin", product_unit_id, delete_permanent)

	daoProduct_unit := p.daoProduct_unit
	if delete_permanent {
		result, err := daoProduct_unit.Delete(product_unit_id)
		if err != nil {
			return err
		}
		log.Printf("Product_unit Service::Delete - End %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(product_unit_id, indata)
		if err != nil {
			return err
		}
		log.Println("Update For Delete Flag", data)
	}
	log.Println("Product_unit Service :: Delete - End")
	return nil
}

func (p *Product_unitBaseService) errorReturn(err error) (Product_unitService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
