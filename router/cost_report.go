package router

import (
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/controller"
	"github.com/songquanpeng/one-api/middleware"
)

// SetupCostReportRoutes 设置费用报表相关路由
func SetupCostReportRoutes(router *gin.Engine) {
	// 费用报表路由组
	costReportGroup := router.Group("/api/cost-report")
	{
		// 身份验证中间件
		costReportGroup.Use(middleware.UserAuth())

		// 按学校维度查询费用报表
		costReportGroup.GET("/school", controller.NewCostReportController().GetCostReportBySchool)

		// 按学科组维度查询费用报表
		costReportGroup.GET("/subject", controller.NewCostReportController().GetCostReportBySubject)

		// 按老师维度查询费用报表
		costReportGroup.GET("/teacher", controller.NewCostReportController().GetCostReportByTeacher)

		// 按用户组维度查询费用报表
		costReportGroup.GET("/user-group", controller.NewCostReportController().GetCostReportByUserGroup)

		// 综合维度费用报表查询
		costReportGroup.GET("/comprehensive", controller.NewCostReportController().GetComprehensiveCostReport)

		// 获取费用报表汇总信息
		costReportGroup.GET("/summary", controller.NewCostReportController().GetCostReportSummary)

		// 获取费用最高的学校列表
		costReportGroup.GET("/top-schools", controller.NewCostReportController().GetTopSchoolsByCost)

		// 获取费用最高的老师列表
		costReportGroup.GET("/top-teachers", controller.NewCostReportController().GetTopTeachersByCost)
	}
}
