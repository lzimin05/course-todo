package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	errs "github.com/lzimin05/course-todo/internal/models/errs"
	models "github.com/lzimin05/course-todo/internal/models/project"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
)

const (
	queryCreateProject = `
		INSERT INTO todo.project (name, description, owner_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at;`

	queryGetProjectByID = `
		SELECT p.id, p.name, p.description, p.owner_id, p.created_at
		FROM todo.project p
		WHERE p.id = $1;`

	queryGetUserProjects = `
		SELECT p.id, p.name, p.description, p.owner_id, p.created_at
		FROM todo.project p
		JOIN todo.project_member pm ON p.id = pm.project_id
		WHERE pm.user_id = $1;`

	queryAddProjectMember = `
		INSERT INTO todo.project_member (project_id, user_id, role)
		VALUES ($1, $2, $3)
		RETURNING id, joined_at;`

	queryGetProjectMembers = `
		SELECT pm.id, pm.project_id, pm.user_id, u.username, u.email, pm.role, pm.joined_at
		FROM todo.project_member pm
		JOIN todo."user" u ON pm.user_id = u.id
		WHERE pm.project_id = $1;`

	queryCheckProjectAccess = `
		SELECT COUNT(*)
		FROM todo.project_member pm
		WHERE pm.project_id = $1 AND pm.user_id = $2;`

	queryDeleteProject = `
		DELETE FROM todo.project
		WHERE id = $1 AND owner_id = $2;`

	queryRemoveProjectMember = `
		DELETE FROM todo.project_member
		WHERE project_id = $1 AND user_id = $2 AND role != 'owner';`

	queryUpdateProject = `
		UPDATE todo.project 
		SET name = $2, description = $3
		WHERE id = $1 AND owner_id = $4;`
)

type ProjectRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) CreateProject(ctx context.Context, project *models.Project) error {
	const op = "ProjectRepository.CreateProject"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		logger.WithError(err).Error("failed to begin transaction")
		return err
	}
	defer tx.Rollback()

	// Создаем проект
	err = tx.QueryRowContext(ctx, queryCreateProject,
		project.Name, project.Description, project.OwnerID).Scan(
		&project.ID, &project.CreatedAt)
	if err != nil {
		logger.WithError(err).Error("failed to create project")
		return err
	}

	// Добавляем владельца как участника
	var memberID uuid.UUID
	var joinedAt interface{}
	err = tx.QueryRowContext(ctx, queryAddProjectMember,
		project.ID, project.OwnerID, models.RoleOwner).Scan(&memberID, &joinedAt)
	if err != nil {
		logger.WithError(err).Error("failed to add owner as member")
		return err
	}

	return tx.Commit()
}

func (r *ProjectRepository) GetProjectByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	const op = "ProjectRepository.GetProjectByID"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	var project models.Project

	err := r.db.QueryRowContext(ctx, queryGetProjectByID, id).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.OwnerID,
		&project.CreatedAt,
	)
	if err != nil {
		logger.WithError(err).Warn("failed to get project by id")
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return &project, nil
}

func (r *ProjectRepository) GetUserProjects(ctx context.Context, userID uuid.UUID) ([]*models.Project, error) {
	const op = "ProjectRepository.GetUserProjects"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	rows, err := r.db.QueryContext(ctx, queryGetUserProjects, userID)
	if err != nil {
		logger.WithError(err).Error("failed to get user projects")
		return nil, err
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		var project models.Project
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.OwnerID,
			&project.CreatedAt,
		)
		if err != nil {
			logger.WithError(err).Error("failed to scan project")
			return nil, err
		}
		projects = append(projects, &project)
	}

	return projects, nil
}

func (r *ProjectRepository) AddProjectMember(ctx context.Context, projectID, userID uuid.UUID) error {
	const op = "ProjectRepository.AddProjectMember"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	var memberID uuid.UUID
	var joinedAt interface{}
	err := r.db.QueryRowContext(ctx, queryAddProjectMember,
		projectID, userID, models.RoleMember).Scan(&memberID, &joinedAt)
	if err != nil {
		logger.WithError(err).Error("failed to add project member")
		return err
	}

	return nil
}

func (r *ProjectRepository) GetProjectMembers(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMember, error) {
	const op = "ProjectRepository.GetProjectMembers"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	rows, err := r.db.QueryContext(ctx, queryGetProjectMembers, projectID)
	if err != nil {
		logger.WithError(err).Error("failed to get project members")
		return nil, err
	}
	defer rows.Close()

	var members []*models.ProjectMember
	for rows.Next() {
		var member models.ProjectMember
		err := rows.Scan(
			&member.ID,
			&member.ProjectID,
			&member.UserID,
			&member.Username,
			&member.Email,
			&member.Role,
			&member.JoinedAt,
		)
		if err != nil {
			logger.WithError(err).Error("failed to scan project member")
			return nil, err
		}
		members = append(members, &member)
	}

	return members, nil
}

func (r *ProjectRepository) CheckProjectAccess(ctx context.Context, projectID, userID uuid.UUID) (bool, error) {
	const op = "ProjectRepository.CheckProjectAccess"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	var count int
	err := r.db.QueryRowContext(ctx, queryCheckProjectAccess, projectID, userID).Scan(&count)
	if err != nil {
		logger.WithError(err).Error("failed to check project access")
		return false, err
	}

	return count > 0, nil
}

func (r *ProjectRepository) DeleteProject(ctx context.Context, projectID, ownerID uuid.UUID) error {
	const op = "ProjectRepository.DeleteProject"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	result, err := r.db.ExecContext(ctx, queryDeleteProject, projectID, ownerID)
	if err != nil {
		logger.WithError(err).Error("failed to delete project")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.WithError(err).Error("failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return errs.ErrNotFound
	}

	return nil
}

func (r *ProjectRepository) RemoveProjectMember(ctx context.Context, projectID, userID uuid.UUID) error {
	const op = "ProjectRepository.RemoveProjectMember"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	result, err := r.db.ExecContext(ctx, queryRemoveProjectMember, projectID, userID)
	if err != nil {
		logger.WithError(err).Error("failed to remove project member")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.WithError(err).Error("failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return errs.ErrNotFound
	}

	return nil
}

func (r *ProjectRepository) UpdateProject(ctx context.Context, projectID uuid.UUID, name, description string, ownerID uuid.UUID) error {
	const op = "ProjectRepository.UpdateProject"
	logger := logctx.GetLogger(ctx).WithField("op", op)

	result, err := r.db.ExecContext(ctx, queryUpdateProject, projectID, name, description, ownerID)
	if err != nil {
		logger.WithError(err).Error("failed to update project")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.WithError(err).Error("failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return errs.ErrNotFound
	}

	return nil
}
