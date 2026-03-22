package services

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/fernandovalenzuela/api-home-pay/internal/repository"
)

type companyService struct {
	repo repository.CompanyRepository
}

func NewCompanyService(repo repository.CompanyRepository) CompanyService {
	return &companyService{repo: repo}
}

func (s *companyService) Create(userID string, company *models.Company) error {
	if err := validateCompany(company); err != nil {
		return err
	}

	exists, err := s.repo.ExistsByName(userID, company.Name)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if exists {
		return fmt.Errorf("company with name '%s' already exists", company.Name)
	}

	return s.repo.Create(userID, company)
}

func (s *companyService) GetByID(userID string, id int) (*models.Company, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid company ID")
	}

	company, err := s.repo.GetByID(userID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}

	return company, nil
}

func (s *companyService) GetAll(userID string) ([]models.Company, error) {
	return s.repo.GetAll(userID)
}

func (s *companyService) Update(userID string, company *models.Company) error {
	if company.ID <= 0 {
		return fmt.Errorf("invalid company ID")
	}

	if err := validateCompany(company); err != nil {
		return err
	}

	existing, err := s.repo.GetByID(userID, company.ID)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("company not found")
	}

	if existing.Name != company.Name {
		exists, err := s.repo.ExistsByName(userID, company.Name)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
		if exists {
			return fmt.Errorf("company with name '%s' already exists", company.Name)
		}
	}

	return s.repo.Update(userID, company)
}

func (s *companyService) Delete(userID string, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid company ID")
	}

	hasServiceAccounts, err := s.repo.HasServiceAccounts(id)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if hasServiceAccounts {
		return fmt.Errorf("cannot delete company with associated service accounts")
	}

	return s.repo.Delete(userID, id)
}

func validateCompany(company *models.Company) error {
	if strings.TrimSpace(company.Name) == "" {
		return fmt.Errorf("company name cannot be empty")
	}

	if company.WebsiteURL != "" {
		if !isValidURL(company.WebsiteURL) {
			return fmt.Errorf("invalid website URL format")
		}
	}

	return nil
}

func isValidURL(urlStr string) bool {
	if urlStr == "" {
		return true
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	if u.Scheme == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return false
	}

	if u.Host == "" {
		return false
	}

	return true
}
