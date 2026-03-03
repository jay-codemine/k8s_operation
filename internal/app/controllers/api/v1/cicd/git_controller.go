package cicd

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
)

// GitController Git 仓库操作控制器
type GitController struct {
}

func NewGitController() *GitController {
	return &GitController{}
}

// GetBranches godoc
// @Summary 获取 Git 仓库分支列表
// @Description 获取指定 Git 仓库的所有远程分支
// @Tags CICD Git
// @Accept json
// @Produce json
// @Param body body requests.GitBranchesRequest true "仓库信息"
// @Success 200 {object} map[string]any "返回分支列表"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cicd/git/branches [post]
func (c *GitController) GetBranches(ctx *gin.Context) {
	param := &requests.GitBranchesRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidGitBranchesRequest); !ok {
		return
	}

	svc := services.NewServices()
	branches, err := svc.GitGetBranches(ctx.Request.Context(), param.RepoURL, param.CredentialID)
	if err != nil {
		global.Logger.Error("GitGetBranches error", zap.Error(err), zap.String("repo", param.RepoURL))
		rsp.ToErrorResponse(errorcode.ErrorGitBranchesFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"branches": branches})
}

// ValidateRepo godoc
// @Summary 验证 Git 仓库连接
// @Description 测试 Git 仓库是否可访问
// @Tags CICD Git
// @Accept json
// @Produce json
// @Param body body requests.GitValidateRequest true "仓库信息"
// @Success 200 {object} map[string]any "验证结果"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cicd/git/validate [post]
func (c *GitController) ValidateRepo(ctx *gin.Context) {
	param := &requests.GitValidateRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidGitValidateRequest); !ok {
		return
	}

	svc := services.NewServices()
	valid, message, err := svc.GitValidateRepo(ctx.Request.Context(), param.RepoURL, param.CredentialID)
	if err != nil {
		global.Logger.Error("GitValidateRepo error", zap.Error(err), zap.String("repo", param.RepoURL))
		rsp.ToErrorResponse(errorcode.ErrorGitValidateFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"valid":   valid,
		"message": message,
	})
}
