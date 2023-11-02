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

type BannerService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(bannerId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(bannerId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(bannerId string, delete_permanent bool) error

	EndService()
}

type bannerBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoBanner   sales_repository.BannerDao
	daoBusiness platform_repository.BusinessDao
	child       BannerService
	businessId  string
}

// NewBannerService - Construct Banner
func NewBannerService(props utils.Map) (BannerService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("BannerService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := bannerBaseService{}
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

// BannerBaseService - Close all the services
func (p *bannerBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *bannerBaseService) initializeService() {
	log.Printf("BannerService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoBanner = sales_repository.NewBannerDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *bannerBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("bannerBaseService::FindAll - Begin")

	listdata, err := p.daoBanner.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("bannerBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *bannerBaseService) Get(bannerId string) (utils.Map, error) {
	log.Printf("bannerBaseService::Get::  Begin %v", bannerId)

	data, err := p.daoBanner.Get(bannerId)

	log.Println("bannerBaseService::Get:: End ", err)
	return data, err
}

func (p *bannerBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("BannerService::FindByCode::  Begin ", filter)

	data, err := p.daoBanner.Find(filter)
	log.Println("BannerService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *bannerBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("BannerService::Create - Begin")
	var bannerId string

	dataval, dataok := indata[sales_common.FLD_BANNER_ID]
	if dataok {
		bannerId = strings.ToLower(dataval.(string))
	} else {
		bannerId = utils.GenerateUniqueId("bnr")
		log.Println("Unique Banner ID", bannerId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_BANNER_ID] = bannerId

	data, err := p.daoBanner.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("BannerService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *bannerBaseService) Update(bannerId string, indata utils.Map) (utils.Map, error) {

	log.Println("BannerService::Update - Begin")

	data, err := p.daoBanner.Update(bannerId, indata)

	log.Println("BannerService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *bannerBaseService) Delete(bannerId string, delete_permanent bool) error {

	log.Println("BannerService::Delete - Begin", bannerId)

	if delete_permanent {
		result, err := p.daoBanner.Delete(bannerId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(bannerId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("BannerService::Delete - End")
	return nil
}

func (p *bannerBaseService) errorReturn(err error) (BannerService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
