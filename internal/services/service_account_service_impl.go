package services

import (
	"fmt"
	"strings"

	"github.com/fernandovalenzuela/api-home-pay/internal/models"
	"github.com/fernandovalenzuela/api-home-pay/internal/repository"
)

type serviceAccountService struct {
	repo repository.ServiceAccountRepository
}

func NewServiceAccountService(repo repository.ServiceAccountRepository) ServiceAccountService {
	return &serviceAccountService{repo: repo}
}

func (s *serviceAccountService) Create(userID string, account *models.ServiceAccount) error {
	if strings.TrimSpace(account.AccountIdentifier) == "" {
		return fmt.Errorf("account identifier cannot be empty")
	}

	if account.CompanyID > 0 {
		exists, err := s.repo.CompanyExistsAndBelongsToUser(userID, account.CompanyID)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
		if !exists {
			return fmt.Errorf("company not found or access denied")
		}

		exists, err = s.repo.ExistsByIdentifier(userID, account.CompanyID, account.AccountIdentifier)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
		if exists {
			return fmt.Errorf("service account with identifier '%s' already exists for this company", account.AccountIdentifier)
		}
	}

	return s.repo.Create(userID, account)
}

func (s *serviceAccountService) GetByID(userID string, id int) (*models.ServiceAccount, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid service account ID")
	}

	account, err := s.repo.GetByID(userID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get service account: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("service account not found")
	}

	return account, nil
}

func (s *serviceAccountService) GetAll(userID string, companyID *int) ([]models.ServiceAccount, error) {
	if companyID != nil && *companyID > 0 {
		exists, err := s.repo.CompanyExistsAndBelongsToUser(userID, *companyID)
		if err != nil {
			return nil, fmt.Errorf("validation error: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("company not found or access denied")
		}
	}

	return s.repo.GetAll(userID, companyID)
}

func (s *serviceAccountService) Update(userID string, account *models.ServiceAccount) error {
	if account.ID <= 0 {
		return fmt.Errorf("invalid service account ID")
	}

	if strings.TrimSpace(account.AccountIdentifier) == "" {
		return fmt.Errorf("account identifier cannot be empty")
	}

	existing, err := s.repo.GetByID(userID, account.ID)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("service account not found")
	}

	if account.CompanyID > 0 {
		exists, err := s.repo.CompanyExistsAndBelongsToUser(userID, account.CompanyID)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
		if !exists {
			return fmt.Errorf("company not found or access denied")
		}

		if existing.AccountIdentifier != account.AccountIdentifier || existing.CompanyID != account.CompanyID {
			exists, err = s.repo.ExistsByIdentifier(userID, account.CompanyID, account.AccountIdentifier)
			if err != nil {
				return fmt.Errorf("validation error: %w", err)
			}
			if exists {
				return fmt.Errorf("service account with identifier '%s' already exists for this company", account.AccountIdentifier)
			}
		}
	}

	return s.repo.Update(userID, account)
}

func (s *serviceAccountService) Delete(userID string, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid service account ID")
	}

	hasExpenses, err := s.repo.HasExpenses(id)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if hasExpenses {
		return fmt.Errorf("cannot delete service account with associated expenses")
	}

	return s.repo.Delete(userID, id)
}
