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

type CampaignService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(campaignId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(campaignId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(campaignId string, delete_permanent bool) error

	EndService()
}

type campaignBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoCampaign sales_repository.CampaignDao
	daoBusiness platform_repository.BusinessDao
	child       CampaignService
	businessId  string
}

// NewCampaignService - Construct Campaign
func NewCampaignService(props utils.Map) (CampaignService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("CampaignService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := campaignBaseService{}
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

// campaignBaseService - Close all the services
func (p *campaignBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *campaignBaseService) initializeService() {
	log.Printf("CampaignService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoCampaign = sales_repository.NewCampaignDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *campaignBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("CampaignBaseService::FindAll - Begin")

	listdata, err := p.daoCampaign.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("CampaignBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *campaignBaseService) Get(campaignId string) (utils.Map, error) {
	log.Printf("campaignBaseService::Get::  Begin %v", campaignId)

	data, err := p.daoCampaign.Get(campaignId)

	log.Println("campaignBaseService::Get:: End ", err)
	return data, err
}

func (p *campaignBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("campaignBaseService::FindByCode::  Begin ", filter)

	data, err := p.daoCampaign.Find(filter)
	log.Println("campaignBaseService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *campaignBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("CampaignService::Create - Begin")
	var campaignId string

	dataval, dataok := indata[sales_common.FLD_CAMPAIGN_ID]
	if dataok {
		campaignId = strings.ToLower(dataval.(string))
	} else {
		campaignId = utils.GenerateUniqueId("camp")
		log.Println("Unique Campaign ID", campaignId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_CAMPAIGN_ID] = campaignId

	data, err := p.daoCampaign.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("CampaignService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *campaignBaseService) Update(campaignId string, indata utils.Map) (utils.Map, error) {

	log.Println("CampaignService::Update - Begin")

	data, err := p.daoCampaign.Update(campaignId, indata)

	log.Println("CampaignService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *campaignBaseService) Delete(campaignId string, delete_permanent bool) error {

	log.Println("CampaignService::Delete - Begin", campaignId)

	if delete_permanent {
		result, err := p.daoCampaign.Delete(campaignId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(campaignId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("CampaignService::Delete - End")
	return nil
}

func (p *campaignBaseService) errorReturn(err error) (CampaignService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
