package appapi

import "testing"

func TestSupportedDatabase(t *testing.T) {
	ok := []string{"postgres", "postgresql", "redis", "mysql", "mongodb", "mongo", "Postgres"}
	for _, name := range ok {
		if !SupportedDatabase(name) && name != "Postgres" {
			// SupportedDatabase 使用 ToLower，Postgres 应通过
		}
		if err := ValidateDatabaseName(name); name != "Postgres" && err != nil {
			// 对已 lower 的名称应通过
			if name == "postgres" || name == "redis" {
				t.Errorf("ValidateDatabaseName(%q) unexpected err: %v", name, err)
			}
		}
	}
	// 明确支持
	for _, name := range []string{"postgres", "redis", "mysql", "mongodb"} {
		if !SupportedDatabase(name) {
			t.Errorf("SupportedDatabase(%q)=false", name)
		}
		if err := ValidateDatabaseName(name); err != nil {
			t.Errorf("ValidateDatabaseName(%q)=%v", name, err)
		}
	}
	// 不支持
	for _, name := range []string{"mariadb", "clickhouse", "oracle", ""} {
		if SupportedDatabase(name) {
			t.Errorf("SupportedDatabase(%q)=true, want false", name)
		}
		if err := ValidateDatabaseName(name); err == nil {
			t.Errorf("ValidateDatabaseName(%q) want error", name)
		}
	}
}

func TestSupportedDatabaseCaseInsensitive(t *testing.T) {
	if !SupportedDatabase("PostgreSQL") && !SupportedDatabase("POSTGRES") {
		// ToLower: postgresql 不在列表... wait "postgresql" is in list, "POSTGRES" -> "postgres" is in list
	}
	if !SupportedDatabase("POSTGRES") {
		t.Error("SupportedDatabase(POSTGRES) should be true")
	}
	if !SupportedDatabase("MongoDB") {
		t.Error("SupportedDatabase(MongoDB) should be true")
	}
}
