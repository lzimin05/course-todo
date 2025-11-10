package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/lzimin05/course-todo/internal/models/errs"
	models "github.com/lzimin05/course-todo/internal/models/project"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/project"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lzimin05/course-todo/internal/usecase/helpers"
)

type ProjectRepository interface {
	CreateProject(ctx context.Context, project *models.Project) error
	GetProjectByID(ctx context.Context, id uuid.UUID) (*models.Project, error)
	GetUserProjects(ctx context.Context, userID uuid.UUID) ([]*models.Project, error)
	AddProjectMember(ctx context.Context, projectID, userID uuid.UUID) error
	GetProjectMembers(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMember, error)
	CheckProjectAccess(ctx context.Context, projectID, userID uuid.UUID) (bool, error)
	DeleteProject(ctx context.Context, projectID, ownerID uuid.UUID) error
	RemoveProjectMember(ctx context.Context, projectID, userID uuid.UUID) error
	UpdateProject(ctx context.Context, projectID uuid.UUID, name, description string, ownerID uuid.UUID) error
}

type ProjectUsecase struct {
	repo ProjectRepository
}

func New(repo ProjectRepository) *ProjectUsecase {
	return &ProjectUsecase{repo: repo}
}

func (uc *ProjectUsecase) CreateProject(ctx context.Context, req *dto.PostProjectDTO) (*dto.ProjectDTO, error) {
	const op = "ProjectUseCase.CreateProject"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("name", req.Name)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return nil, err
	}

	newProject := &models.Project{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     userID,
	}

	err = uc.repo.CreateProject(ctx, newProject)
	if err != nil {
		logger.WithError(err).Error("failed to create project")
		return nil, err
	}

	return &dto.ProjectDTO{
		ID:          newProject.ID,
		Name:        newProject.Name,
		Description: newProject.Description,
		OwnerID:     newProject.OwnerID,
		CreatedAt:   newProject.CreatedAt,
	}, nil
}

func (uc *ProjectUsecase) GetUserProjects(ctx context.Context) ([]*dto.ProjectDTO, error) {
	const op = "ProjectUseCase.GetUserProjects"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return nil, err
	}

	projects, err := uc.repo.GetUserProjects(ctx, userID)
	if err != nil {
		logger.WithError(err).Error("failed to get user projects")
		return nil, err
	}

	projectDTOs := make([]*dto.ProjectDTO, len(projects))
	for i, project := range projects {
		projectDTOs[i] = &dto.ProjectDTO{
			ID:          project.ID,
			Name:        project.Name,
			Description: project.Description,
			OwnerID:     project.OwnerID,
			CreatedAt:   project.CreatedAt,
		}
	}

	return projectDTOs, nil
}

func (uc *ProjectUsecase) GetProjectByID(ctx context.Context, projectID uuid.UUID) (*dto.ProjectDTO, error) {
	const op = "ProjectUseCase.GetProjectByID"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("projectID", projectID)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return nil, err
	}

	// Проверяем доступ к проекту
	hasAccess, err := uc.repo.CheckProjectAccess(ctx, projectID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to check project access")
		return nil, err
	}

	if !hasAccess {
		logger.Warn("user doesn't have access to project")
		return nil, errs.ErrNoAccess
	}

	project, err := uc.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		logger.WithError(err).Error("failed to get project")
		return nil, err
	}

	return &dto.ProjectDTO{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		OwnerID:     project.OwnerID,
		CreatedAt:   project.CreatedAt,
	}, nil
}

func (uc *ProjectUsecase) AddProjectMember(ctx context.Context, projectID uuid.UUID, req *dto.AddMemberDTO) error {
	const op = "ProjectUseCase.AddProjectMember"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("projectID", projectID)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return err
	}

	// Проверяем, что пользователь является владельцем проекта
	project, err := uc.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		logger.WithError(err).Error("failed to get project")
		return err
	}

	if project.OwnerID != userID {
		logger.Warn("user is not project owner")
		return errs.ErrNotOwner
	}

	err = uc.repo.AddProjectMember(ctx, projectID, req.UserID)
	if err != nil {
		logger.WithError(err).Error("failed to add project member")
		return err
	}

	return nil
}

func (uc *ProjectUsecase) GetProjectMembers(ctx context.Context, projectID uuid.UUID) ([]*dto.ProjectMemberDTO, error) {
	const op = "ProjectUseCase.GetProjectMembers"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("projectID", projectID)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return nil, err
	}

	// Проверяем доступ к проекту
	hasAccess, err := uc.repo.CheckProjectAccess(ctx, projectID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to check project access")
		return nil, err
	}

	if !hasAccess {
		logger.Warn("user doesn't have access to project")
		return nil, errs.ErrNoAccess
	}

	members, err := uc.repo.GetProjectMembers(ctx, projectID)
	if err != nil {
		logger.WithError(err).Error("failed to get project members")
		return nil, err
	}

	memberDTOs := make([]*dto.ProjectMemberDTO, len(members))
	for i, member := range members {
		memberDTOs[i] = &dto.ProjectMemberDTO{
			ID:        member.ID,
			ProjectID: member.ProjectID,
			UserID:    member.UserID,
			Role:      member.Role,
			JoinedAt:  member.JoinedAt,
		}
	}

	return memberDTOs, nil
}

func (uc *ProjectUsecase) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	const op = "ProjectUseCase.DeleteProject"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("projectID", projectID)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return err
	}

	err = uc.repo.DeleteProject(ctx, projectID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to delete project")
		return err
	}

	return nil
}

func (uc *ProjectUsecase) RemoveProjectMember(ctx context.Context, projectID, memberUserID uuid.UUID) error {
	const op = "ProjectUseCase.RemoveProjectMember"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("projectID", projectID)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return err
	}

	// Проверяем, что пользователь является владельцем проекта
	project, err := uc.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		logger.WithError(err).Error("failed to get project")
		return err
	}

	if project.OwnerID != userID {
		logger.Warn("user is not project owner")
		return errs.ErrNotOwner
	}

	err = uc.repo.RemoveProjectMember(ctx, projectID, memberUserID)
	if err != nil {
		logger.WithError(err).Error("failed to remove project member")
		return err
	}

	return nil
}

func (uc *ProjectUsecase) UpdateProject(ctx context.Context, projectID uuid.UUID, req *dto.UpdateProjectDTO) (*dto.ProjectDTO, error) {
	const op = "ProjectUseCase.UpdateProject"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("projectID", projectID)

	userID, err := helpers.GetUserIDFromContext(ctx)
	if err != nil {
		logger.WithError(err).Error("invalid user ID format")
		return nil, err
	}

	// Проверяем, что пользователь является владельцем проекта
	project, err := uc.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		logger.WithError(err).Error("failed to get project")
		return nil, err
	}

	if project.OwnerID != userID {
		logger.Warn("user is not project owner")
		return nil, errs.ErrNotOwner
	}

	err = uc.repo.UpdateProject(ctx, projectID, req.Name, req.Description, userID)
	if err != nil {
		logger.WithError(err).Error("failed to update project")
		return nil, err
	}

	updatedProject, err := uc.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		logger.WithError(err).Error("failed to get updated project")
		return nil, err
	}

	return &dto.ProjectDTO{
		ID:          updatedProject.ID,
		Name:        updatedProject.Name,
		Description: updatedProject.Description,
		OwnerID:     updatedProject.OwnerID,
		CreatedAt:   updatedProject.CreatedAt,
	}, nil
}
