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

type OfferService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(offerId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(offerId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(offerId string, delete_permanent bool) error

	EndService()
}

type offerBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoOffer    sales_repository.OfferDao
	daoBusiness platform_repository.BusinessDao
	child       OfferService
	businessId  string
}

// NewOfferService - Construct Offer
func NewOfferService(props utils.Map) (OfferService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("OfferService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := offerBaseService{}
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

// offersBaseService - Close all the services
func (p *offerBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *offerBaseService) initializeService() {
	log.Printf("OfferService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoOffer = sales_repository.NewOfferDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *offerBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("offerBaseService::FindAll - Begin")

	listdata, err := p.daoOffer.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("offerBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *offerBaseService) Get(offerId string) (utils.Map, error) {
	log.Printf("offerBaseService::Get::  Begin %v", offerId)

	data, err := p.daoOffer.Get(offerId)

	log.Println("offerBaseService::Get:: End ", err)
	return data, err
}

func (p *offerBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("OfferService::FindByCode::  Begin ", filter)

	data, err := p.daoOffer.Find(filter)
	log.Println("OfferService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *offerBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("OfferService::Create - Begin")
	var offerId string

	dataval, dataok := indata[sales_common.FLD_OFFER_ID]
	if dataok {
		offerId = strings.ToLower(dataval.(string))
	} else {
		offerId = utils.GenerateUniqueId("offr")
		log.Println("Unique Offer ID", offerId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_OFFER_ID] = offerId

	data, err := p.daoOffer.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("OfferService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *offerBaseService) Update(offerId string, indata utils.Map) (utils.Map, error) {

	log.Println("OfferService::Update - Begin")

	data, err := p.daoOffer.Update(offerId, indata)

	log.Println("OfferService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *offerBaseService) Delete(offerId string, delete_permanent bool) error {

	log.Println("OfferService::Delete - Begin", offerId)

	if delete_permanent {
		result, err := p.daoOffer.Delete(offerId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(offerId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("OfferService::Delete - End")
	return nil
}

func (p *offerBaseService) errorReturn(err error) (OfferService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
