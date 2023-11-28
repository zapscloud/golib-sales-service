package customer_service

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

type CustomerCartService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(cartId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(cartId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(cartId string, delete_permanent bool) error

	EndService()
}

type customerCartBaseService struct {
	db_utils.DatabaseService
	dbRegion        db_utils.DatabaseService
	daoCustomerCart customer_repository.CustomerCartDao
	daoBusiness     platform_repository.BusinessDao
	daoCustomer     sales_repository.CustomerDao

	child      CustomerCartService
	businessId string
	customerId string
}

// NewCustomerCartService - Construct CustomerCart
func NewCustomerCartService(props utils.Map) (CustomerCartService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("CustomerCartService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := customerCartBaseService{}
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

// customerCartBaseService - Close all the service
func (p *customerCartBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *customerCartBaseService) initializeService() {
	log.Printf("CustomerCartService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoCustomer = sales_repository.NewCustomerDao(p.dbRegion.GetClient(), p.businessId)
	p.daoCustomerCart = customer_repository.NewCustomerCartDao(p.dbRegion.GetClient(), p.businessId, p.customerId)
}

// List - List All records
func (p *customerCartBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("customerCartBaseService::FindAll - Begin")

	listdata, err := p.daoCustomerCart.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("customerCartBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *customerCartBaseService) Get(cartId string) (utils.Map, error) {
	log.Printf("customerCartBaseService::Get::  Begin %v", cartId)

	data, err := p.daoCustomerCart.Get(cartId)

	log.Println("customerCartBaseService::Get:: End ", err)
	return data, err
}

func (p *customerCartBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("CustomerCartService::FindByCode::  Begin ", filter)

	data, err := p.daoCustomerCart.Find(filter)
	log.Println("CustomerCartService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *customerCartBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("CustomerCartService::Create - Begin")
	var cartId string

	dataval, dataok := indata[sales_common.FLD_CART_ID]
	if dataok {
		cartId = strings.ToLower(dataval.(string))
	} else {
		cartId = utils.GenerateUniqueId("crt")
		log.Println("Unique CustomerCart ID", cartId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_CUSTOMER_ID] = p.customerId
	indata[sales_common.FLD_CART_ID] = cartId

	data, err := p.daoCustomerCart.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("CustomerCartService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *customerCartBaseService) Update(cartId string, indata utils.Map) (utils.Map, error) {

	log.Println("CustomerCartService::Update - Begin")

	// Delete Key values
	delete(indata, sales_common.FLD_BUSINESS_ID)
	delete(indata, sales_common.FLD_CUSTOMER_ID)
	delete(indata, sales_common.FLD_CART_ID)

	data, err := p.daoCustomerCart.Update(cartId, indata)

	log.Println("CustomerCartService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *customerCartBaseService) Delete(cartId string, delete_permanent bool) error {

	log.Println("CustomerCartService::Delete - Begin", cartId)

	if delete_permanent {
		result, err := p.daoCustomerCart.Delete(cartId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(cartId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("CustomerCartService::Delete - End")
	return nil
}

func (p *customerCartBaseService) errorReturn(err error) (CustomerCartService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
