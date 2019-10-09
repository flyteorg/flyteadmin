package transformers

import (
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

type CreateProjectModelInput struct {
	Identifier  string
	Name        string
	Description string
}

func CreateProjectModel(identifier, name string, description string) models.Project {
	return models.Project{
		Identifier:  identifier,
		Name:        name,
		Description: description,
	}
}

func FromProjectModel(projectModel models.Project, domains []*admin.Domain) admin.Project {
	project := admin.Project{
		Id:          projectModel.Identifier,
		Name:        projectModel.Name,
		Description: projectModel.Description,
	}
	project.Domains = domains
	return project
}

func FromProjectModels(projectModels []models.Project, domains []*admin.Domain) []*admin.Project {
	projects := make([]*admin.Project, len(projectModels))
	for index, projectModel := range projectModels {
		project := FromProjectModel(projectModel, domains)
		projects[index] = &project
	}
	return projects
}
