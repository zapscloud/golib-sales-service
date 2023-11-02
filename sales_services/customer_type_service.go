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

type CustomerTypeService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(CustomerTypeId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(CustomerTypeId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(CustomerTypeId string, delete_permanent bool) error

	EndService()
}

type CustomerTypeBaseService struct {
	db_utils.DatabaseService
	dbRegion        db_utils.DatabaseService
	daoCustomerType sales_repository.CustomerTypeDao
	daoBusiness     platform_repository.BusinessDao
	child           CustomerTypeService
	businessId      string
}

// NewCustomerTypeService - Construct CustomerType
func NewCustomerTypeService(props utils.Map) (CustomerTypeService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("CustomerTypeService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := CustomerTypeBaseService{}
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

// EndService - Close all the services
func (p *CustomerTypeBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *CustomerTypeBaseService) initializeService() {
	log.Printf("CustomerTypeService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoCustomerType = sales_repository.NewCustomerTypeDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *CustomerTypeBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("CustomerTypeBaseService::FindAll - Begin")

	listdata, err := p.daoCustomerType.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("CustomerTypeBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *CustomerTypeBaseService) Get(CustomerTypeId string) (utils.Map, error) {
	log.Printf("CustomerTypeBaseService::Get::  Begin %v", CustomerTypeId)

	data, err := p.daoCustomerType.Get(CustomerTypeId)
	if err != nil {
		return nil, err
	}
	log.Println("CustomerTypeBaseService::Get:: End ", err)
	return data, err
}

func (p *CustomerTypeBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("CustomerTypeService::FindByCode::  Begin ", filter)

	data, err := p.daoCustomerType.Find(filter)
	log.Println("CustomerTypeService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *CustomerTypeBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("CustomerTypeService::Create - Begin")
	var CustomerTypeId string

	dataval, dataok := indata[sales_common.FLD_CUSTOMER_TYPE_ID]
	if dataok {
		CustomerTypeId = strings.ToLower(dataval.(string))
	} else {
		CustomerTypeId = utils.GenerateUniqueId("cust")
		log.Println("Unique CustomerType ID", CustomerTypeId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_CUSTOMER_TYPE_ID] = CustomerTypeId

	data, err := p.daoCustomerType.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("CustomerTypeService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *CustomerTypeBaseService) Update(CustomerTypeId string, indata utils.Map) (utils.Map, error) {

	log.Println("CustomerTypeService::Update - Begin")

	// Delete the Key fields if exist
	delete(indata, sales_common.FLD_BUSINESS_ID)
	delete(indata, sales_common.FLD_CUSTOMER_TYPE_ID)

	data, err := p.daoCustomerType.Update(CustomerTypeId, indata)

	log.Println("CustomerTypeService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *CustomerTypeBaseService) Delete(CustomerTypeId string, delete_permanent bool) error {

	log.Println("CustomerTypeService::Delete - Begin", CustomerTypeId)

	if delete_permanent {
		result, err := p.daoCustomerType.Delete(CustomerTypeId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(CustomerTypeId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("CustomerTypeService::Delete - End")
	return nil
}

func (p *CustomerTypeBaseService) errorReturn(err error) (CustomerTypeService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
