package routes

import (
	"task-manager/controller"
	attachmentcontroller "task-manager/controller/attachmentController"
	"task-manager/controller/commentController"
	"task-manager/controller/labelController"
	"task-manager/controller/taskController"
	"task-manager/controller/teamController"
	"task-manager/controller/userController"
	"task-manager/middleware"

	"github.com/gin-gonic/gin"
)

func SetUpRoutes(router *gin.Engine) {
	router.GET("/", controller.Home)

	public := router.Group("/api")
	{
		public.POST("/register", userController.Register)
		public.POST("/login", userController.Login)
	}
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware())

	user_router := protected.Group("/users")
	{
		user_router.GET("/", userController.GetMyDetails)
		user_router.GET("/:id", userController.GetUserById)
		user_router.DELETE("/", userController.DeleteMyAccount)
		user_router.DELETE("/:id", userController.DeleteUserById)
		user_router.GET("/all-users", userController.Users)
		user_router.PUT("/", userController.UpdateMyAccount)
		user_router.PUT("/:id", userController.UpdateUserById)
		user_router.PUT("/:id/role", userController.UpdateUserRole)
		user_router.PUT("/change-password", userController.UpdateMyPassword)
	}

	team_router := protected.Group("/teams")
	{
		team_router.POST("/", teamController.RegisterTeam)
		team_router.GET("/", teamController.GetAllTeams)
		team_router.GET("/my-teams", teamController.GetMyTeams)
		team_router.GET("/:id", teamController.GetTeamById)
		team_router.PUT("/:id", teamController.UpdateTeam)
		team_router.PUT("/:id/add-members", teamController.AddMembersInTeam)
		team_router.PUT("/:id/remove-members", teamController.RemoveMembersFromTeam)
		team_router.DELETE("/:id", teamController.DeleteTeamById)
	}

	task_router := protected.Group("/tasks")
	{
		task_router.POST("/", taskController.CreateTask)
		task_router.GET("/", taskController.GetAllTasks)
		task_router.GET("/assigned-tasks", taskController.GetMyAssignedTasks)
		task_router.GET("/created-tasks", taskController.GetMyCreatedTasks)
		task_router.GET("/:id", taskController.GetTaskById)
		task_router.PUT("/:id", taskController.UpdateTaskById)
		task_router.PUT("/:id/change-team", taskController.UpdateTaskTeamById)
		task_router.PUT("/:id/add-assignee", taskController.AddTaskAssignees)
		task_router.PUT("/:id/remove-assignee", taskController.RemoveTaskAssignees)
		task_router.PUT("/:id/parent-id", taskController.UpdateParentTaskId)
		task_router.DELETE("/:id", taskController.DeleteTaskById)

		// Task - Labels
		task_router.PUT("/:id/add-labels", taskController.AddTaskLabels)
		task_router.PUT("/:id/remove-labels", taskController.RemoveTaskLabels)
		task_router.GET("/:id/labels", taskController.GetAllLabelsOfTask)

		// Task - Comments
		task_router.POST("/:id/comments", taskController.AddCommentToTask)
		task_router.GET("/:id/comments", taskController.GetAllCommentsOfTask)

		// Task - Attachment
		task_router.POST("/:id/attachments", taskController.UploadAttachment)
		task_router.GET("/:id/attachments", taskController.GetAllTaskAttachments)
	}

	label_router := protected.Group("/labels")
	{
		label_router.POST("/", labelController.CreateLabel)
		label_router.GET("/", labelController.GetAllLabels)
		label_router.GET("/:id", labelController.GetLabelById)
		label_router.PUT("/:id", labelController.UpdateLabelById)
		label_router.DELETE("/:id", labelController.DeleteLabelById)
	}

	comment_router := protected.Group("/comments")
	{
		comment_router.GET("/", commentController.GetAllComments)
		comment_router.GET("/:id", commentController.GetCommentById)
		comment_router.GET("/my-comments", commentController.GetMyComments)
		comment_router.DELETE("/:id", commentController.DeleteCommentById)
		comment_router.PUT("/:id", commentController.UpdateCommentById)
	}

	attachment_router := protected.Group("/attachments")
	{
		attachment_router.GET("/:id", attachmentcontroller.GetAttachmentById)
		attachment_router.DELETE("/:id", attachmentcontroller.DeleteAttachmentById)
	}
}
