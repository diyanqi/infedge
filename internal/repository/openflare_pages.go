// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"context"

	db "github.com/Rain-kl/Wavelet/internal/infra/persistence"
	"github.com/Rain-kl/Wavelet/internal/model"
)

// HasPagesProjectsTable 判断 Pages 项目表是否已迁移。
func HasPagesProjectsTable(ctx context.Context) bool {
	return db.DB(ctx).Migrator().HasTable(&model.PagesProject{})
}

// ListPagesProjects 列出全部 Pages 项目。
func ListPagesProjects(ctx context.Context) ([]model.PagesProject, error) {
	var projects []model.PagesProject
	if err := db.DB(ctx).Order("id desc").Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

// ListPagesProjectsByOwner lists only projects owned by one user.
func ListPagesProjectsByOwner(ctx context.Context, ownerID uint64) ([]model.PagesProject, error) {
	var projects []model.PagesProject
	if err := db.DB(ctx).Where("owner_id = ?", ownerID).Order("id desc").Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

// GetPagesProjectByID 按 ID 查询 Pages 项目。
func GetPagesProjectByID(ctx context.Context, id uint) (*model.PagesProject, error) {
	var project model.PagesProject
	if err := db.DB(ctx).First(&project, id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// GetPagesProjectByIDAndOwner scopes a project to its owner.
func GetPagesProjectByIDAndOwner(ctx context.Context, id uint, ownerID uint64) (*model.PagesProject, error) {
	var project model.PagesProject
	if err := db.DB(ctx).Where("id = ? AND owner_id = ?", id, ownerID).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// GetPagesProjectBySlug 按 slug 查询 Pages 项目。
func GetPagesProjectBySlug(ctx context.Context, slug string) (*model.PagesProject, error) {
	var project model.PagesProject
	if err := db.DB(ctx).Where("slug = ?", slug).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// CreatePagesProjectRecord 创建 Pages 项目。
func CreatePagesProjectRecord(ctx context.Context, project *model.PagesProject) error {
	return db.DB(ctx).Create(project).Error
}

// ListPagesDeployments 列出项目的全部部署。
func ListPagesDeployments(ctx context.Context, projectID uint) ([]model.PagesDeployment, error) {
	var deployments []model.PagesDeployment
	if err := db.DB(ctx).Where("project_id = ?", projectID).Order("id desc").Find(&deployments).Error; err != nil {
		return nil, err
	}
	return deployments, nil
}

// GetPagesDeploymentByID 按 ID 查询 Pages 部署。
func GetPagesDeploymentByID(ctx context.Context, id uint) (*model.PagesDeployment, error) {
	var deployment model.PagesDeployment
	if err := db.DB(ctx).First(&deployment, id).Error; err != nil {
		return nil, err
	}
	return &deployment, nil
}

// GetPagesProjectByDeploymentIDAndOwner verifies a deployment belongs to an owner's project.
func GetPagesProjectByDeploymentIDAndOwner(ctx context.Context, deploymentID uint, ownerID uint64) (*model.PagesProject, error) {
	var project model.PagesProject
	if err := db.DB(ctx).Joins("JOIN of_pages_deployments ON of_pages_deployments.project_id = of_pages_projects.id").
		Where("of_pages_deployments.id = ? AND of_pages_projects.owner_id = ?", deploymentID, ownerID).
		First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// ListPagesDeploymentFiles 列出部署文件清单。
func ListPagesDeploymentFiles(ctx context.Context, deploymentID uint) ([]model.PagesDeploymentFile, error) {
	var files []model.PagesDeploymentFile
	if err := db.DB(ctx).Where("deployment_id = ?", deploymentID).Order("path asc").Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

// CountPagesDeploymentsByProjectID 统计项目部署数量。
func CountPagesDeploymentsByProjectID(ctx context.Context, projectID uint) (int64, error) {
	var count int64
	if err := db.DB(ctx).Model(&model.PagesDeployment{}).Where("project_id = ?", projectID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountProxyRoutesByPagesProjectID 统计引用 Pages 项目的代理规则数量。
func CountProxyRoutesByPagesProjectID(ctx context.Context, projectID uint) (int64, error) {
	if !HasProxyRoutesTable(ctx) {
		return 0, nil
	}
	var count int64
	if err := db.DB(ctx).Model(&model.ProxyRoute{}).Where("pages_project_id = ?", projectID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
