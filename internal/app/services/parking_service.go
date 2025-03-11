package services

import (
	"github.com/agilistikmal/parkingo-core/internal/app/models"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ParkingService struct {
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewParkingService(db *gorm.DB, validate *validator.Validate) *ParkingService {
	return &ParkingService{
		DB:       db,
		Validate: validate,
	}
}

func (s *ParkingService) GetParkings() ([]models.Parking, error) {
	var parkings []models.Parking
	err := s.DB.Find(&parkings).Error
	if err != nil {
		return nil, err
	}

	return parkings, nil
}

func (s *ParkingService) GetParkingByID(id int) (*models.Parking, error) {
	var parking *models.Parking
	err := s.DB.First(&parking, id).Error
	if err != nil {
		return nil, err
	}

	return parking, nil
}

func (s *ParkingService) GetParkingSlotsByParkingID(parkingID int) ([]models.ParkingSlot, error) {
	var slots []models.ParkingSlot
	err := s.DB.Where("parking_id = ?", parkingID).Find(&slots).Error
	if err != nil {
		return nil, err
	}

	return slots, nil
}

func (s *ParkingService) GetParkingSlotByID(id int) (*models.ParkingSlot, error) {
	var slot *models.ParkingSlot
	err := s.DB.First(&slot, id).Error
	if err != nil {
		return nil, err
	}

	return slot, nil
}

func (s *ParkingService) CreateParking(authorID int, req *models.CreateParkingRequest) (*models.Parking, error) {
	err := s.Validate.Struct(req)
	if err != nil {
		return nil, err
	}

	parking := models.Parking{
		AuthorID:  authorID,
		Name:      req.Name,
		Address:   req.Address,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Layout:    req.Layout,
	}

	err = s.DB.Create(&parking).Error
	if err != nil {
		return nil, err
	}

	return &parking, nil
}

func (s *ParkingService) UpdateParking(id int, req *models.UpdateParkingRequest) (*models.Parking, error) {
	err := s.Validate.Struct(req)
	if err != nil {
		return nil, err
	}

	parking, err := s.GetParkingByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		parking.Name = req.Name
	}
	if req.Address != "" {
		parking.Address = req.Address
	}
	if req.Latitude != 0 {
		parking.Latitude = req.Latitude
	}
	if req.Longitude != 0 {
		parking.Longitude = req.Longitude
	}
	if req.Layout != nil {
		parking.Layout = req.Layout
	}

	err = s.DB.Save(&parking).Error
	if err != nil {
		return nil, err
	}

	return parking, nil
}

func (s *ParkingService) DeleteParking(id int) error {
	parking, err := s.GetParkingByID(id)
	if err != nil {
		return err
	}

	err = s.DB.Delete(&parking).Error
	if err != nil {
		return err
	}
	return nil
}

func (s *ParkingService) CreateParkingSlot(req *models.CreateParkingSlotRequest) (*models.ParkingSlot, error) {
	err := s.Validate.Struct(req)
	if err != nil {
		return nil, err
	}

	slot := models.ParkingSlot{
		ParkingID: req.ParkingID,
		Name:      req.Name,
		Status:    req.Status,
		Fee:       req.Fee,
		Row:       req.Col,
		Col:       req.Row,
	}

	err = s.DB.Create(&slot).Error
	if err != nil {
		return nil, err
	}

	return &slot, nil
}

func (s *ParkingService) UpdateParkingSlot(id int, req *models.UpdateParkingSlotRequest) (*models.ParkingSlot, error) {
	err := s.Validate.Struct(req)
	if err != nil {
		return nil, err
	}

	slot, err := s.GetParkingSlotByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		slot.Name = req.Name
	}
	if req.Status != "" {
		slot.Status = req.Status
	}
	if req.Fee != 0 {
		slot.Fee = req.Fee
	}
	if req.Row != 0 {
		slot.Row = req.Row
	}
	if req.Col != 0 {
		slot.Col = req.Col
	}

	err = s.DB.Save(&slot).Error
	if err != nil {
		return nil, err
	}

	return slot, nil
}

func (s *ParkingService) DeleteParkingSlot(id int) error {
	slot, err := s.GetParkingSlotByID(id)
	if err != nil {
		return err
	}

	err = s.DB.Delete(&slot).Error
	if err != nil {
		return err
	}

	return nil
}
