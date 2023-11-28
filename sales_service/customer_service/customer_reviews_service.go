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

type CustomerReviewService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(reviewId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(reviewId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(reviewId string, delete_permanent bool) error

	EndService()
}

type customerReviewBaseService struct {
	db_utils.DatabaseService
	dbRegion          db_utils.DatabaseService
	daoCustomerReview customer_repository.CustomerReviewDao
	daoBusiness       platform_repository.BusinessDao
	daoCustomer       sales_repository.CustomerDao

	child      CustomerReviewService
	businessId string
	customerId string
}

// NewCustomerReviewService - Construct CustomerReview
func NewCustomerReviewService(props utils.Map) (CustomerReviewService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("CustomerReviewService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := customerReviewBaseService{}
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

// customerReviewBaseService - Close all the services
func (p *customerReviewBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *customerReviewBaseService) initializeService() {
	log.Printf("CustomerReviewService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoCustomer = sales_repository.NewCustomerDao(p.dbRegion.GetClient(), p.businessId)
	p.daoCustomerReview = customer_repository.NewCustomerReviewDao(p.GetClient(), p.businessId, p.customerId)
}

// List - List All records
func (p *customerReviewBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("customerReviewBaseService::FindAll - Begin")

	listdata, err := p.daoCustomerReview.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("customerReviewBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *customerReviewBaseService) Get(reviewId string) (utils.Map, error) {
	log.Printf("customerReviewBaseService::Get::  Begin %v", reviewId)

	data, err := p.daoCustomerReview.Get(reviewId)

	log.Println("customerReviewBaseService::Get:: End ", err)
	return data, err
}

func (p *customerReviewBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("customerReviewBaseService::FindByCode::  Begin ", filter)

	data, err := p.daoCustomerReview.Find(filter)
	log.Println("customerReviewBaseService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *customerReviewBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("CustomerReviewService::Create - Begin")
	var reviewId string

	dataval, dataok := indata[sales_common.FLD_REVIEW_ID]
	if dataok {
		reviewId = strings.ToLower(dataval.(string))
	} else {
		reviewId = utils.GenerateUniqueId("reviw")
		log.Println("Unique CustomerReview ID", reviewId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_CUSTOMER_ID] = p.customerId
	indata[sales_common.FLD_REVIEW_ID] = reviewId

	data, err := p.daoCustomerReview.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("CustomerReviewService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *customerReviewBaseService) Update(reviewId string, indata utils.Map) (utils.Map, error) {

	log.Println("CustomerReviewService::Update - Begin")

	// Delete Key values
	delete(indata, sales_common.FLD_BUSINESS_ID)
	delete(indata, sales_common.FLD_CUSTOMER_ID)
	delete(indata, sales_common.FLD_REVIEW_ID)

	data, err := p.daoCustomerReview.Update(reviewId, indata)

	log.Println("CustomerReviewService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *customerReviewBaseService) Delete(reviewId string, delete_permanent bool) error {

	log.Println("CustomerReviewService::Delete - Begin", reviewId)

	if delete_permanent {
		result, err := p.daoCustomerReview.Delete(reviewId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(reviewId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("CustomerReviewService::Delete - End")
	return nil
}

func (p *customerReviewBaseService) errorReturn(err error) (CustomerReviewService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
