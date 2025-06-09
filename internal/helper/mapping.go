package helper

import (
	"fmt"
	"strings"
	"time"

	hrPb "github.com/longgggwwww/hrm-ms-hr/ent/proto/entpb"
	permPb "github.com/longgggwwww/hrm-ms-permission/ent/proto/entpb"
)

func GetProtoValue(field interface{}) string {
	if field == nil {
		return ""
	}
	if v, ok := field.(interface{ GetValue() string }); ok {
		return v.GetValue()
	}
	if m, ok := field.(map[string]interface{}); ok {
		if val, ok := m["value"].(string); ok {
			return val
		}
	}
	return ""
}

// Helper: convert protobuf Timestamp to RFC3339 string
func ProtoTimestampToString(ts interface{}) string {
	t, ok := ts.(interface {
		GetSeconds() int64
		GetNanos() int32
	})
	if !ok || t.GetSeconds() == 0 {
		return ""
	}
	tm := time.Unix(t.GetSeconds(), int64(t.GetNanos())).UTC()
	return tm.Format(time.RFC3339)
}

// Helper: map roles to clean array (no id, no proto fields, formatted time)
func ToRoleArr(roles []*permPb.RoleExt) []map[string]interface{} {
	arr := make([]map[string]interface{}, 0, len(roles))
	for _, r := range roles {
		// map perms as full perm objects (not just code)
		perms := make([]map[string]interface{}, 0, len(r.Perms))
		for _, p := range r.Perms {
			perms = append(perms, map[string]interface{}{
				"code":        p.Code,
				"name":        p.Name,
				"description": GetProtoValue(p.Description),
			})
		}
		arr = append(arr, map[string]interface{}{
			"code":        r.Code,
			"name":        r.Name,
			"color":       GetProtoValue(r.Color),
			"description": GetProtoValue(r.Description),
			"perms":       perms,
			"created_at":  ProtoTimestampToString(r.CreatedAt),
			"updated_at":  ProtoTimestampToString(r.UpdatedAt),
		})
	}
	return arr
}

// Helper: map perms to clean array (no id, no proto fields)
func ToPermArr(perms []*permPb.PermExt) []map[string]interface{} {
	arr := make([]map[string]interface{}, 0, len(perms))
	for _, p := range perms {
		arr = append(arr, map[string]interface{}{
			"code":        p.Code,
			"name":        p.Name,
			"description": GetProtoValue(p.Description),
		})
	}
	return arr
}

func toOrganizationMap(o *hrPb.Organization) map[string]interface{} {
	if o == nil {
		return nil
	}
	return map[string]interface{}{
		"id":   o.Id,
		"name": o.Name,
		"code": o.Code,
	}
}

func toDepartmentMap(d *hrPb.Department) map[string]interface{} {
	if d == nil {
		return nil
	}
	return map[string]interface{}{
		"id":           d.Id,
		"name":         d.Name,
		"code":         d.Code,
		"organization": toOrganizationMap(d.Organization),
	}
}

func toPositionMap(p *hrPb.Position) map[string]interface{} {
	if p == nil {
		return nil
	}
	return map[string]interface{}{
		"id":          p.Id,
		"name":        p.Name,
		"code":        p.Code,
		"departments": toDepartmentMap(p.Departments),
	}
}

func parseStatusEnum(status interface{}) string {
	statusStr := ""
	switch v := status.(type) {
	case string:
		statusStr = v
	case fmt.Stringer:
		statusStr = v.String()
	default:
		return ""
	}
	parts := strings.Split(statusStr, "_")
	if len(parts) > 1 {
		return strings.ToLower(parts[1])
	}
	return strings.ToLower(statusStr)
}

// Helper: convert hrPb.Employee to map[string]interface{}
func ToEmployeeMap(e *hrPb.Employee) map[string]interface{} {
	if e == nil {
		return nil
	}
	return map[string]interface{}{
		"id":          e.Id,
		"user_id":     GetProtoValue(e.UserId),
		"code":        e.Code,
		"status":      parseStatusEnum(e.Status),
		"position_id": e.PositionId,
		"joining_at":  ProtoTimestampToString(e.JoiningAt),
		"org_id":      e.OrgId,
		"created_at":  ProtoTimestampToString(e.CreatedAt),
		"updated_at":  ProtoTimestampToString(e.UpdatedAt),
		"position":    toPositionMap(e.Position),
	}
}
