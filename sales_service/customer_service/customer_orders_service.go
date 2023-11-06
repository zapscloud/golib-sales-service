package customer_services

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
	"github.com/zapscloud/golib-sales-repository/sales_repository/customer_repository"
	"github.com/zapscloud/golib-utils/utils"
)

type CustomerOrderService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(customerOrderId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(bcustomerOrderId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(customerOrderId string, delete_permanent bool) error

	EndService()
}

type customerOrderBaseService struct {
	db_utils.DatabaseService
	dbRegion         db_utils.DatabaseService
	daoCustomerOrder customer_repository.CustomerOrderDao
	daoBusiness      platform_repository.BusinessDao
	daoCustomer      sales_repository.CustomerDao

	child      CustomerOrderService
	businessId string
	customerId string
}

// NewCustomerOrderService - Construct CustomerOrder
func NewCustomerOrderService(props utils.Map) (CustomerOrderService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("CustomerOrderService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := customerOrderBaseService{}
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

	// Verify whether the User id data passed, this is optional parameter
	customerId, _ := utils.GetMemberDataStr(props, sales_common.FLD_CUSTOMER_ID)
	// if err != nil {
	// 	return p.errorReturn(err)
	// }

	// Assign the BusinessId
	p.businessId = businessId
	p.customerId = customerId
	p.initializeService()

	// Verify the Business Exists
	_, err = p.daoBusiness.Get(businessId)
	if err != nil {
		err := &utils.AppError{
			ErrorCode:   funcode + "01",
			ErrorMsg:    "Invalid BusinessId",
			ErrorDetail: "Given BusinessId is not exist"}
		return p.errorReturn(err)
	}

	// Verify the Customer Exist
	if len(customerId) > 0 {
		_, err = p.daoCustomer.Get(customerId)
		if err != nil {
			err := &utils.AppError{
				ErrorCode:   funcode + "01",
				ErrorMsg:    "Invalid CustomerId",
				ErrorDetail: "Given CustomerId is not exist"}
			return p.errorReturn(err)
		}
	}

	p.child = &p

	return &p, err
}

// customerOrderBaseService - Close all the services
func (p *customerOrderBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *customerOrderBaseService) initializeService() {
	log.Printf("customerOrderBaseService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoCustomer = sales_repository.NewCustomerDao(p.dbRegion.GetClient(), p.businessId)
	p.daoCustomerOrder = customer_repository.NewCustomerOrderDao(p.GetClient(), p.businessId, p.customerId)
}

// List - List All records
func (p *customerOrderBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("customerOrderBaseService::FindAll - Begin")

	listdata, err := p.daoCustomerOrder.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("customerOrderBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *customerOrderBaseService) Get(custOrderId string) (utils.Map, error) {
	log.Printf("customerOrderBaseService::Get::  Begin %v", custOrderId)

	data, err := p.daoCustomerOrder.Get(custOrderId)

	log.Println("customerOrderBaseService::Get:: End ", err)
	return data, err
}

func (p *customerOrderBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("customerOrderBaseService::FindByCode::  Begin ", filter)

	data, err := p.daoCustomerOrder.Find(filter)
	log.Println("customerOrderBaseService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *customerOrderBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("customerOrderBaseService::Create - Begin")
	var custOrderId string

	dataval, dataok := indata[sales_common.FLD_CUSTOMER_ORDER_ID]
	if dataok {
		custOrderId = strings.ToLower(dataval.(string))
	} else {
		custOrderId = utils.GenerateUniqueId("cust_order")
		log.Println("Unique customerOrder ID", custOrderId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_CUSTOMER_ID] = p.customerId
	indata[sales_common.FLD_CUSTOMER_ORDER_ID] = custOrderId

	data, err := p.daoCustomerOrder.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("customerOrderBaseService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *customerOrderBaseService) Update(custOrderId string, indata utils.Map) (utils.Map, error) {

	log.Println("customerOrderService::Update - Begin")

	// Delete Key values
	delete(indata, sales_common.FLD_BUSINESS_ID)
	delete(indata, sales_common.FLD_CUSTOMER_ID)
	delete(indata, sales_common.FLD_CUSTOMER_ORDER_ID)

	data, err := p.daoCustomerOrder.Update(custOrderId, indata)

	log.Println("customerOrderService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *customerOrderBaseService) Delete(custOrderId string, delete_permanent bool) error {

	log.Println("customerOrderService::Delete - Begin", custOrderId)

	if delete_permanent {
		result, err := p.daoCustomerOrder.Delete(custOrderId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(custOrderId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("customerOrderService::Delete - End")
	return nil
}

func (p *customerOrderBaseService) errorReturn(err error) (CustomerOrderService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
