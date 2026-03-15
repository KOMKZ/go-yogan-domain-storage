package permissions

type DeclaredPermission struct {
	PermissionCode string
	PermissionName string
	PermissionType string
	ResourceCode   string
	GroupCode      string
	Description    string
}

func DeclaredPermissions() []DeclaredPermission {
	return []DeclaredPermission{
		{
			PermissionCode: "file:read",
			PermissionName: "查看文件",
			PermissionType: "READ",
			ResourceCode:   "file",
			GroupCode:      "SYSTEM",
			Description:    "文件列表、详情、下载",
		},
		{
			PermissionCode: "file:write",
			PermissionName: "管理文件",
			PermissionType: "WRITE",
			ResourceCode:   "file",
			GroupCode:      "SYSTEM",
			Description:    "文件上传、删除与更新",
		},
	}
}
