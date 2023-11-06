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

type CustomerWishlistService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(wishlistId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(wishlistId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(wishlistId string, delete_permanent bool) error

	EndService()
}

type customerWishlistBaseService struct {
	db_utils.DatabaseService
	dbRegion            db_utils.DatabaseService
	daoCustomerWishlist customer_repository.CustomerWishlistDao
	daoBusiness         platform_repository.BusinessDao
	daoCustomer         sales_repository.CustomerDao

	child      CustomerWishlistService
	businessId string
	customerId string
}

// NewCustomerWishlistService - Construct CustomerWishlist
func NewCustomerWishlistService(props utils.Map) (CustomerWishlistService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("CustomerWishlistService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := customerWishlistBaseService{}
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
	// 	return nil, err
	// }

	// Assign the BusinessId
	p.businessId = businessId
	p.customerId = customerId
	p.initializeService()

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

// customerWishlistBaseService - Close all the services
func (p *customerWishlistBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *customerWishlistBaseService) initializeService() {
	log.Printf("CustomerWishlistService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoCustomer = sales_repository.NewCustomerDao(p.dbRegion.GetClient(), p.businessId)
	p.daoCustomerWishlist = customer_repository.NewCustomerWishlistDao(p.GetClient(), p.businessId, p.customerId)
}

// List - List All records
func (p *customerWishlistBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("customerWishlistBaseService::FindAll - Begin")

	listdata, err := p.daoCustomerWishlist.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("customerWishlistBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *customerWishlistBaseService) Get(wishlistId string) (utils.Map, error) {
	log.Printf("customerWishlistBaseService::Get::  Begin %v", wishlistId)

	data, err := p.daoCustomerWishlist.Get(wishlistId)

	log.Println("customerWishlistBaseService::Get:: End ", err)
	return data, err
}

func (p *customerWishlistBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("customerWishlistBaseService::FindByCode::  Begin ", filter)

	data, err := p.daoCustomerWishlist.Find(filter)
	log.Println("customerWishlistBaseService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *customerWishlistBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("CustomerWishlistService::Create - Begin")
	var wishlistId string

	dataval, dataok := indata[sales_common.FLD_WISHLIST_ID]
	if dataok {
		wishlistId = strings.ToLower(dataval.(string))
	} else {
		wishlistId = utils.GenerateUniqueId("wish")
		log.Println("Unique CustomerWishlist ID", wishlistId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_CUSTOMER_ID] = p.customerId
	indata[sales_common.FLD_WISHLIST_ID] = wishlistId

	data, err := p.daoCustomerWishlist.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("CustomerWishlistService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *customerWishlistBaseService) Update(wishlistId string, indata utils.Map) (utils.Map, error) {

	log.Println("CustomerWishlistService::Update - Begin")

	// Delete Key values
	delete(indata, sales_common.FLD_BUSINESS_ID)
	delete(indata, sales_common.FLD_CUSTOMER_ID)
	delete(indata, sales_common.FLD_WISHLIST_ID)

	data, err := p.daoCustomerWishlist.Update(wishlistId, indata)

	log.Println("CustomerWishlistService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *customerWishlistBaseService) Delete(wishlistId string, delete_permanent bool) error {

	log.Println("CustomerWishlistService::Delete - Begin", wishlistId)

	if delete_permanent {
		result, err := p.daoCustomerWishlist.Delete(wishlistId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(wishlistId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("CustomerWishlistService::Delete - End")
	return nil
}

func (p *customerWishlistBaseService) errorReturn(err error) (CustomerWishlistService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
