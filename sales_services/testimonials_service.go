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

type TestimonialService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(testimonialId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(testimonialId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(testimonialId string, delete_permanent bool) error

	EndService()
}

type testimonialBaseService struct {
	db_utils.DatabaseService
	dbRegion       db_utils.DatabaseService
	daoTestimonial sales_repository.TestimonialDao
	daoBusiness    platform_repository.BusinessDao
	child          TestimonialService
	businessId     string
}

// NewTestimonialService - Construct Testimonail
func NewTestimonialService(props utils.Map) (TestimonialService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("TestimonialService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := testimonialBaseService{}
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

// testmonialBaseService - Close all the services
func (p *testimonialBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *testimonialBaseService) initializeService() {
	log.Printf("TestimonialService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoTestimonial = sales_repository.NewTestimonialDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *testimonialBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("testimonialBaseService::FindAll - Begin")

	listdata, err := p.daoTestimonial.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("testimonialBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *testimonialBaseService) Get(testimonialId string) (utils.Map, error) {
	log.Printf("testimonialBaseService::Get::  Begin %v", testimonialId)

	data, err := p.daoTestimonial.Get(testimonialId)

	log.Println("testimonialBaseService::Get:: End ", err)
	return data, err
}

func (p *testimonialBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("TestimonialService::FindByCode::  Begin ", filter)

	data, err := p.daoTestimonial.Find(filter)
	log.Println("TestimonialService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *testimonialBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("TestimonialService::Create - Begin")
	var testimonialId string

	dataval, dataok := indata[sales_common.FLD_TESTIMONIAL_ID]
	if dataok {
		testimonialId = strings.ToLower(dataval.(string))
	} else {
		testimonialId = utils.GenerateUniqueId("tes")
		log.Println("Unique Testimonial ID", testimonialId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_TESTIMONIAL_ID] = testimonialId

	data, err := p.daoTestimonial.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("TestimonialService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *testimonialBaseService) Update(testimonialId string, indata utils.Map) (utils.Map, error) {

	log.Println("TestimonialService::Update - Begin")

	data, err := p.daoTestimonial.Update(testimonialId, indata)

	log.Println("TestimonialService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *testimonialBaseService) Delete(testimonialId string, delete_permanent bool) error {

	log.Println("TestimonialService::Delete - Begin", testimonialId)

	if delete_permanent {
		result, err := p.daoTestimonial.Delete(testimonialId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(testimonialId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("TestimonialService::Delete - End")
	return nil
}

func (p *testimonialBaseService) errorReturn(err error) (TestimonialService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
