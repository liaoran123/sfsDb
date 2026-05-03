package engine

import (
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// 批量添加测试
func TestBatchInsertNoInc_BothStrategies(t *testing.T) {
	t.Run("FailFast_MissingPrimaryKey", func(t *testing.T) {
		table, err := NewTable("test_batch_fail_fast_1")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"id": 1, "name": "Alice", "age": 25},
			{"name": "Bob", "age": 30},
			{"id": 3, "name": "Charlie", "age": 35},
		}

		ids, err := table.BatchInsertNoInc(records, false)
		if err == nil {
			t.Fatal("expected error when record missing primary key, got nil")
		}
		if ids != nil && len(ids) > 0 {
			t.Errorf("expected nil ids on error, got %v", ids)
		}
		t.Logf("FailFast correctly returned error: %v", err)
	})

	t.Run("Continue_MissingPrimaryKey", func(t *testing.T) {
		table, err := NewTable("test_batch_continue_1")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"id": 1, "name": "Alice", "age": 25},
			{"name": "Bob", "age": 30},
			{"id": 3, "name": "Charlie", "age": 35},
		}

		ids, err := table.BatchInsertNoInc(records, true)
		if err != nil {
			t.Fatalf("Continue strategy should not return error: %v", err)
		}
		if len(ids) != 2 {
			t.Errorf("expected 2 ids, got %d", len(ids))
		}
	})

	t.Run("FailFast_NilPrimaryKey", func(t *testing.T) {
		table, err := NewTable("test_batch_fail_fast_2")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"id": 1, "name": "Alice", "age": 25},
			{"id": nil, "name": "Bob", "age": 30},
			{"id": 3, "name": "Charlie", "age": 35},
		}

		ids, err := table.BatchInsertNoInc(records, false)
		if err == nil {
			t.Fatal("expected error when record has nil primary key, got nil")
		}
		t.Logf("FailFast correctly returned error for nil pk: %v", err)
		if ids != nil && len(ids) > 0 {
			t.Errorf("expected nil ids on error, got %v", ids)
		}
	})

	t.Run("Continue_NilPrimaryKey", func(t *testing.T) {
		table, err := NewTable("test_batch_continue_2")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"id": 1, "name": "Alice", "age": 25},
			{"id": nil, "name": "Bob", "age": 30},
			{"id": 3, "name": "Charlie", "age": 35},
		}

		ids, err := table.BatchInsertNoInc(records, true)
		if err != nil {
			t.Fatalf("Continue strategy should not return error: %v", err)
		}
		if len(ids) != 2 {
			t.Errorf("expected 2 ids, got %d", len(ids))
		}
	})

	t.Run("AllValid_NoDifferenceBetweenStrategies", func(t *testing.T) {
		table1, err := NewTable("test_all_valid_1")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}
		table2, err := NewTable("test_all_valid_2")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0}
		table1.SetFields(fields)
		table2.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table1.CreateIndex(pk)
		table2.CreateIndex(pk)

		records := []*map[string]any{
			{"id": 1, "name": "Alice", "age": 25},
			{"id": 2, "name": "Bob", "age": 30},
			{"id": 3, "name": "Charlie", "age": 35},
		}

		ids1, err1 := table1.BatchInsertNoInc(records, false)
		ids2, err2 := table2.BatchInsertNoInc(records, true)

		if err1 != nil {
			t.Fatalf("FailFast should not return error for all valid records: %v", err1)
		}
		if err2 != nil {
			t.Fatalf("Continue should not return error for all valid records: %v", err2)
		}
		if len(ids1) != len(ids2) || len(ids1) != 3 {
			t.Errorf("expected 3 ids from both strategies, got %d and %d", len(ids1), len(ids2))
		}
	})

	t.Run("EmptyRecords", func(t *testing.T) {
		table, err := NewTable("test_empty_records")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{}

		ids, err := table.BatchInsertNoInc(records, false)
		if err != nil {
			t.Fatalf("should not return error for empty records: %v", err)
		}
		if len(ids) != 0 {
			t.Errorf("expected 0 ids for empty records, got %d", len(ids))
		}
	})

	t.Run("DataIntegrityAfterContinue", func(t *testing.T) {
		table, err := NewTable("test_data_integrity")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"id": 1, "name": "Alice", "age": 25},
			{"name": "INVALID"}, // 跳过
			{"id": 3, "name": "Charlie", "age": 35},
		}

		ids, err := table.BatchInsertNoInc(records, true)
		if err != nil {
			t.Fatalf("Continue strategy failed: %v", err)
		}
		if len(ids) != 2 {
			t.Fatalf("expected 2 ids, got %d", len(ids))
		}

		for _, id := range ids {
			searchFields := map[string]any{"id": id}
			iter, err := table.Search(&searchFields)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}
			if iter == nil {
				t.Fatal("iterator is nil")
			}
			defer GlobalTableIterPool.Put(iter)
			if !iter.Next() {
				t.Errorf("record with id %d not found", id)
				continue
			}
			recs := iter.GetRecords(true)
			if len(recs) != 1 {
				t.Errorf("expected 1 record for id %d, got %d", id, len(recs))
			}
		}
	})
}

func TestBatchInsertNoIncIoT_BothStrategies(t *testing.T) {
	t.Run("FailFast_WithSyncOption", func(t *testing.T) {
		table, err := NewTable("test_iot_fail_fast")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"id": 1, "name": "Alice", "age": 25},
			{"name": "Bob", "age": 30},
			{"id": 3, "name": "Charlie", "age": 35},
		}

		ids, err := table.BatchInsertNoIncIoT(records, false)
		if err == nil {
			t.Fatal("expected error with FailFast")
		}
		if ids != nil && len(ids) > 0 {
			t.Errorf("expected nil ids on error, got %v", ids)
		}
	})

	t.Run("Continue_WithSyncOption", func(t *testing.T) {
		table, err := NewTable("test_iot_continue")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"id": 1, "name": "Alice", "age": 25},
			{"name": "Bob", "age": 30},
			{"id": 3, "name": "Charlie", "age": 35},
		}

		ids, err := table.BatchInsertNoIncIoT(records, true)
		if err != nil {
			t.Fatalf("Continue strategy should not return error: %v", err)
		}
		if len(ids) != 2 {
			t.Errorf("expected 2 ids, got %d", len(ids))
		}
	})

	t.Run("AllValid_DataPersisted", func(t *testing.T) {
		table, err := NewTable("test_iot_all_valid")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"id": 101, "name": "IoT_Alice", "age": 25},
			{"id": 102, "name": "IoT_Bob", "age": 30},
			{"id": 103, "name": "IoT_Charlie", "age": 35},
		}

		ids, err := table.BatchInsertNoIncIoT(records, false)
		if err != nil {
			t.Fatalf("BatchInsertNoIncIoT failed: %v", err)
		}
		if len(ids) != 3 {
			t.Errorf("expected 3 ids, got %d", len(ids))
		}

		for i, id := range ids {
			expectedID := (*records[i])["id"]
			if id != expectedID {
				t.Errorf("id mismatch: expected %v, got %v", expectedID, id)
			}
		}
	})
}

func TestBatchInsertNoIncImpl_ContinueStrategy_ImplLevel(t *testing.T) {
	t.Run("Impl_BatchInsertNoInc", func(t *testing.T) {
		table, err := NewTable("test_impl_batch_no_inc")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"id": 1, "name": "Alice", "age": 25},
			{"name": "INVALID"},
			{"id": 3, "name": "Charlie", "age": 35},
		}

		impl := NewBatchInsertImpl(table, records)
		ids, err := impl.BatchInsertNoInc(true)
		if err != nil {
			t.Fatalf("BatchInsertNoInc(true) failed: %v", err)
		}
		if len(ids) != 2 {
			t.Errorf("expected 2 ids, got %d", len(ids))
		}
	})

	t.Run("Impl_BatchInsertNoIncIoT", func(t *testing.T) {
		table, err := NewTable("test_impl_batch_no_inc_iot")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"id": 1, "name": "Alice", "age": 25},
			{"name": "INVALID"},
			{"id": 3, "name": "Charlie", "age": 35},
		}

		impl := NewBatchInsertImpl(table, records)
		ids, err := impl.BatchInsertNoIncIoT(true)
		if err != nil {
			t.Fatalf("BatchInsertNoIncIoT(true) failed: %v", err)
		}
		if len(ids) != 2 {
			t.Errorf("expected 2 ids, got %d", len(ids))
		}
	})
}

func TestBatchInsertInc_BothStrategies(t *testing.T) {
	t.Run("FailFast_InvalidRecord", func(t *testing.T) {
		table, err := NewTable("test_batch_insert_inc_failfast")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"name": "Alice", "age": 25},
			{"name": "Bob", "age": "invalid"}, // 无效字段类型
			{"name": "Charlie", "age": 35},
		}

		_, testErr := table.BatchInsertInc(records, false)
		if testErr == nil {
			t.Fatal("expected error when invalid record with FailFast strategy, but got nil")
		}
		t.Logf("FailFast correctly returned error: %v", testErr)
	})

	t.Run("Continue_InvalidRecord", func(t *testing.T) {
		table, err := NewTable("test_batch_insert_inc_continue")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"name": "Alice", "age": 25},
			{"name": "Bob", "age": "invalid"}, // 无效字段类型 - 应跳过
			{"name": "Charlie", "age": 35},
		}

		ids, err := table.BatchInsertInc(records, true)
		if err != nil {
			t.Fatalf("Continue strategy should not return error: %v", err)
		}

		if len(ids) != 2 {
			t.Errorf("expected 2 valid IDs, got %d: %v", len(ids), ids)
		}

		for _, id := range ids {
			searchFields := map[string]any{"id": id}
			iter, err := table.Search(&searchFields)
			if err != nil {
				t.Fatalf("Search failed for id %d: %v", id, err)
			}
			if iter == nil {
				t.Fatalf("Failed to get iterator for id %d", id)
			}
			defer GlobalTableIterPool.Put(iter)

			if !iter.Next() {
				t.Errorf("record with id %d not found", id)
			}
		}
		t.Logf("Continue strategy test passed, inserted IDs: %v", ids)
	})

	t.Run("AllValid_NoDifference", func(t *testing.T) {
		table1, err := NewTable("test_batch_insert_inc_all_valid1")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}
		table2, err := NewTable("test_batch_insert_inc_all_valid2")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0}
		table1.SetFields(fields)
		table2.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table1.CreateIndex(pk)
		table2.CreateIndex(pk)

		records := []*map[string]any{
			{"name": "Alice", "age": 25},
			{"name": "Bob", "age": 30},
			{"name": "Charlie", "age": 35},
		}

		ids1, err1 := table1.BatchInsertInc(records, false)
		ids2, err2 := table2.BatchInsertInc(records, true)

		if err1 != nil {
			t.Fatalf("FailFast should not error on valid records: %v", err1)
		}
		if err2 != nil {
			t.Fatalf("Continue should not error on valid records: %v", err2)
		}

		if len(ids1) != 3 || len(ids2) != 3 {
			t.Errorf("both strategies should insert 3 records, got %d and %d", len(ids1), len(ids2))
		}
	})
}

func TestBatchInsertNoInc_DataCorrectness(t *testing.T) {
	t.Run("ContinueStrategy_VerifyAllFields", func(t *testing.T) {
		table, err := NewTable("test_data_correctness")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0, "email": "", "active": false}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"id": 1, "name": "Alice", "age": 25, "email": "alice@example.com", "active": true},
			{"id": 2, "name": "Bob", "age": 30, "email": "bob@example.com", "active": false},
			{"id": 3, "name": "Charlie", "age": 35, "email": "charlie@example.com", "active": true},
		}

		ids, err := table.BatchInsertNoInc(records, true)
		if err != nil {
			t.Fatalf("BatchInsertNoInc failed: %v", err)
		}

		if len(ids) != 3 {
			t.Fatalf("expected 3 ids, got %d", len(ids))
		}

		for i, id := range ids {
			searchFields := map[string]any{"id": id}
			iter, err := table.Search(&searchFields)
			if err != nil {
				t.Fatalf("Search failed for id %d: %v", id, err)
			}
			if iter == nil {
				t.Fatalf("Failed to get iterator for id %d", id)
			}
			defer GlobalTableIterPool.Put(iter)

			if !iter.Next() {
				t.Fatalf("record with id %d not found", id)
			}

			resultRecords := iter.GetRecords(false)
			if len(resultRecords) != 1 {
				t.Fatalf("expected 1 record for id %d, got %d", id, len(resultRecords))
			}

			result := resultRecords[0]
			expected := records[i]

			if result["name"] != (*expected)["name"] {
				t.Errorf("id %d: name mismatch, expected %v, got %v", id, (*expected)["name"], result["name"])
			}
			if result["age"] != (*expected)["age"] {
				t.Errorf("id %d: age mismatch, expected %v, got %v", id, (*expected)["age"], result["age"])
			}
			if result["email"] != (*expected)["email"] {
				t.Errorf("id %d: email mismatch, expected %v, got %v", id, (*expected)["email"], result["email"])
			}
			if result["active"] != (*expected)["active"] {
				t.Errorf("id %d: active mismatch, expected %v, got %v", id, (*expected)["active"], result["active"])
			}
		}
		t.Logf("All fields verified correctly for all inserted records")
	})

	t.Run("ContinueStrategy_SkipInvalid_VerifyValid", func(t *testing.T) {
		table, err := NewTable("test_skip_invalid_verify_valid")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"id": 100, "name": "Valid1", "age": 20, "score": 95.5},
			{"name": "INVALID"},
			{"id": 101, "name": "Valid2", "age": 25, "score": 88.0},
			{"id": nil, "name": "INVALID"},
			{"id": 102, "name": "Valid3", "age": 30, "score": 91.25},
		}

		ids, err := table.BatchInsertNoInc(records, true)
		if err != nil {
			t.Fatalf("Continue strategy failed: %v", err)
		}

		if len(ids) != 3 {
			t.Fatalf("expected 3 valid IDs, got %d: %v", len(ids), ids)
		}

		expectedRecords := []*map[string]any{
			{"id": 100, "name": "Valid1", "age": 20, "score": 95.5},
			{"id": 101, "name": "Valid2", "age": 25, "score": 88.0},
			{"id": 102, "name": "Valid3", "age": 30, "score": 91.25},
		}

		for i, id := range ids {
			searchFields := map[string]any{"id": id}
			iter, err := table.Search(&searchFields)
			if err != nil {
				t.Fatalf("Search failed for id %d: %v", id, err)
			}
			if iter == nil {
				t.Fatalf("Failed to get iterator for id %d", id)
			}
			defer GlobalTableIterPool.Put(iter)

			if !iter.Next() {
				t.Fatalf("record with id %d not found", id)
			}

			resultRecords := iter.GetRecords(false)
			if len(resultRecords) != 1 {
				t.Fatalf("expected 1 record for id %d, got %d", id, len(resultRecords))
			}

			result := resultRecords[0]
			expected := expectedRecords[i]

			if result["name"] != (*expected)["name"] {
				t.Errorf("id %d: name mismatch, expected %v, got %v", id, (*expected)["name"], result["name"])
			}
			if result["age"] != (*expected)["age"] {
				t.Errorf("id %d: age mismatch, expected %v, got %v", id, (*expected)["age"], result["age"])
			}
			if result["score"] != (*expected)["score"] {
				t.Errorf("id %d: score mismatch, expected %v, got %v", id, (*expected)["score"], result["score"])
			}
		}
		t.Logf("Skip invalid verify valid test passed, all valid records have correct data")
	})
}

func TestBatchInsertInc_DataCorrectness(t *testing.T) {
	t.Run("AutoID_VerifyAllFields", func(t *testing.T) {
		table, err := NewTable("test_auto_id_correctness")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0, "email": ""}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"name": "User1", "age": 18, "email": "user1@test.com"},
			{"name": "User2", "age": 22, "email": "user2@test.com"},
			{"name": "User3", "age": 28, "email": "user3@test.com"},
		}

		ids, err := table.BatchInsertInc(records, true)
		if err != nil {
			t.Fatalf("BatchInsertInc failed: %v", err)
		}

		if len(ids) != 3 {
			t.Fatalf("expected 3 ids, got %d", len(ids))
		}

		for i, id := range ids {
			searchFields := map[string]any{"id": id}
			iter, err := table.Search(&searchFields)
			if err != nil {
				t.Fatalf("Search failed for id %d: %v", id, err)
			}
			if iter == nil {
				t.Fatalf("Failed to get iterator for id %d", id)
			}
			defer GlobalTableIterPool.Put(iter)

			if !iter.Next() {
				t.Fatalf("record with id %d not found", id)
			}

			resultRecords := iter.GetRecords(false)
			if len(resultRecords) != 1 {
				t.Fatalf("expected 1 record for id %d, got %d", id, len(resultRecords))
			}

			result := resultRecords[0]
			expected := records[i]

			if result["name"] != (*expected)["name"] {
				t.Errorf("id %d: name mismatch, expected %v, got %v", id, (*expected)["name"], result["name"])
			}
			if result["age"] != (*expected)["age"] {
				t.Errorf("id %d: age mismatch, expected %v, got %v", id, (*expected)["age"], result["age"])
			}
			if result["email"] != (*expected)["email"] {
				t.Errorf("id %d: email mismatch, expected %v, got %v", id, (*expected)["email"], result["email"])
			}
		}
		t.Logf("Auto ID test passed, all fields verified correctly")
	})

	t.Run("AutoID_SkipInvalid_VerifyValid", func(t *testing.T) {
		table, err := NewTable("test_auto_id_skip_invalid")
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"name": "ValidFirst", "age": 21},
			{"name": "InvalidAge", "age": "not_a_number"},
			{"name": "ValidSecond", "age": 22},
		}

		ids, err := table.BatchInsertInc(records, true)
		if err != nil {
			t.Fatalf("Continue strategy should not return error: %v", err)
		}

		if len(ids) != 2 {
			t.Fatalf("expected 2 valid IDs, got %d: %v", len(ids), ids)
		}

		expectedRecords := []*map[string]any{
			{"name": "ValidFirst", "age": 21},
			{"name": "ValidSecond", "age": 22},
		}

		for i, id := range ids {
			searchFields := map[string]any{"id": id}
			iter, err := table.Search(&searchFields)
			if err != nil {
				t.Fatalf("Search failed for id %d: %v", id, err)
			}
			if iter == nil {
				t.Fatalf("Failed to get iterator for id %d", id)
			}
			defer GlobalTableIterPool.Put(iter)

			if !iter.Next() {
				t.Fatalf("record with id %d not found", id)
			}

			resultRecords := iter.GetRecords(false)
			if len(resultRecords) != 1 {
				t.Fatalf("expected 1 record for id %d, got %d", id, len(resultRecords))
			}

			result := resultRecords[0]
			expected := expectedRecords[i]

			if result["name"] != (*expected)["name"] {
				t.Errorf("id %d: name mismatch, expected %v, got %v", id, (*expected)["name"], result["name"])
			}
			if result["age"] != (*expected)["age"] {
				t.Errorf("id %d: age mismatch, expected %v, got %v", id, (*expected)["age"], result["age"])
			}
		}
		t.Logf("Auto ID skip invalid verify valid test passed")
	})
}

func TestBatchInsert_Encrypted(t *testing.T) {
	t.Run("Encrypted_BatchInsertNoInc_VerifyAllFields", func(t *testing.T) {
		masterKey := make([]byte, 32)
		for i := range masterKey {
			masterKey[i] = byte(i)
		}

		encryptConfig := &storage.EncryptionConfig{
			Enabled:   true,
			Algorithm: "AES-256-GCM",
			MasterKey: masterKey,
		}

		dbPath := "./test_encrypted_batch_noinc_db_" + t.Name()
		tableName := "test_encrypted_batch_noinc_" + t.Name()

		_, err := storage.GetDBManager().OpenDB(dbPath, encryptConfig)
		if err != nil {
			t.Fatalf("Failed to open encrypted database: %v", err)
		}

		table, err := NewTable(tableName)
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0, "email": "", "active": false}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"id": 1, "name": "Alice", "age": 25, "email": "alice@encrypted.com", "active": true},
			{"id": 2, "name": "Bob", "age": 30, "email": "bob@encrypted.com", "active": false},
			{"id": 3, "name": "Charlie", "age": 35, "email": "charlie@encrypted.com", "active": true},
		}

		ids, err := table.BatchInsertNoInc(records, false)
		if err != nil {
			t.Fatalf("Encrypted BatchInsertNoInc failed: %v", err)
		}

		if len(ids) != 3 {
			t.Fatalf("expected 3 ids, got %d", len(ids))
		}

		for i, id := range ids {
			searchFields := map[string]any{"id": id}
			iter, err := table.Search(&searchFields)
			if err != nil {
				t.Fatalf("Search failed for id %d: %v", id, err)
			}
			if iter == nil {
				t.Fatalf("Failed to get iterator for id %d", id)
			}
			defer GlobalTableIterPool.Put(iter)

			if !iter.Next() {
				t.Fatalf("record with id %d not found", id)
			}

			resultRecords := iter.GetRecords(false)
			if len(resultRecords) != 1 {
				t.Fatalf("expected 1 record for id %d, got %d", id, len(resultRecords))
			}

			result := resultRecords[0]
			expected := records[i]

			if result["name"] != (*expected)["name"] {
				t.Errorf("id %d: name mismatch in encrypted DB, expected %v, got %v", id, (*expected)["name"], result["name"])
			}
			if result["age"] != (*expected)["age"] {
				t.Errorf("id %d: age mismatch in encrypted DB, expected %v, got %v", id, (*expected)["age"], result["age"])
			}
			if result["email"] != (*expected)["email"] {
				t.Errorf("id %d: email mismatch in encrypted DB, expected %v, got %v", id, (*expected)["email"], result["email"])
			}
			if result["active"] != (*expected)["active"] {
				t.Errorf("id %d: active mismatch in encrypted DB, expected %v, got %v", id, (*expected)["active"], result["active"])
			}
		}
		t.Logf("Encrypted BatchInsertNoInc test passed, all fields verified correctly")
	})

	t.Run("Encrypted_BatchInsertInc_VerifyAllFields", func(t *testing.T) {
		masterKey := make([]byte, 32)
		for i := range masterKey {
			masterKey[i] = byte(i)
		}

		encryptConfig := &storage.EncryptionConfig{
			Enabled:   true,
			Algorithm: "AES-256-GCM",
			MasterKey: masterKey,
		}

		dbPath := "./test_encrypted_batch_inc_db_" + t.Name()
		tableName := "test_encrypted_batch_inc_" + t.Name()

		_, err := storage.GetDBManager().OpenDB(dbPath, encryptConfig)
		if err != nil {
			t.Fatalf("Failed to open encrypted database: %v", err)
		}

		table, err := NewTable(tableName)
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0, "email": ""}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"name": "User1", "age": 18, "email": "user1@encrypted.com"},
			{"name": "User2", "age": 22, "email": "user2@encrypted.com"},
			{"name": "User3", "age": 28, "email": "user3@encrypted.com"},
		}

		ids, err := table.BatchInsertInc(records, false)
		if err != nil {
			t.Fatalf("Encrypted BatchInsertInc failed: %v", err)
		}

		if len(ids) != 3 {
			t.Fatalf("expected 3 ids, got %d", len(ids))
		}

		for i, id := range ids {
			searchFields := map[string]any{"id": id}
			iter, err := table.Search(&searchFields)
			if err != nil {
				t.Fatalf("Search failed for id %d: %v", id, err)
			}
			if iter == nil {
				t.Fatalf("Failed to get iterator for id %d", id)
			}
			defer GlobalTableIterPool.Put(iter)

			if !iter.Next() {
				t.Fatalf("record with id %d not found", id)
			}

			resultRecords := iter.GetRecords(false)
			if len(resultRecords) != 1 {
				t.Fatalf("expected 1 record for id %d, got %d", id, len(resultRecords))
			}

			result := resultRecords[0]
			expected := records[i]

			if result["name"] != (*expected)["name"] {
				t.Errorf("id %d: name mismatch in encrypted DB, expected %v, got %v", id, (*expected)["name"], result["name"])
			}
			if result["age"] != (*expected)["age"] {
				t.Errorf("id %d: age mismatch in encrypted DB, expected %v, got %v", id, (*expected)["age"], result["age"])
			}
			if result["email"] != (*expected)["email"] {
				t.Errorf("id %d: email mismatch in encrypted DB, expected %v, got %v", id, (*expected)["email"], result["email"])
			}
		}
		t.Logf("Encrypted BatchInsertInc test passed, all fields verified correctly")
	})

	t.Run("Encrypted_BatchInsertNoIncIoT_VerifyAllFields", func(t *testing.T) {
		masterKey := make([]byte, 32)
		for i := range masterKey {
			masterKey[i] = byte(i)
		}

		encryptConfig := &storage.EncryptionConfig{
			Enabled:   true,
			Algorithm: "AES-256-GCM",
			MasterKey: masterKey,
		}

		dbPath := "./test_encrypted_batch_iot_db_" + t.Name()
		tableName := "test_encrypted_batch_iot_" + t.Name()

		_, err := storage.GetDBManager().OpenDB(dbPath, encryptConfig)
		if err != nil {
			t.Fatalf("Failed to open encrypted database: %v", err)
		}

		table, err := NewTable(tableName)
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0, "email": ""}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"id": 1, "name": "IoT_Device1", "age": 25, "email": "device1@iot.com"},
			{"id": 2, "name": "IoT_Device2", "age": 30, "email": "device2@iot.com"},
			{"id": 3, "name": "IoT_Device3", "age": 35, "email": "device3@iot.com"},
		}

		ids, err := table.BatchInsertNoIncIoT(records, false)
		if err != nil {
			t.Fatalf("Encrypted BatchInsertNoIncIoT failed: %v", err)
		}

		if len(ids) != 3 {
			t.Fatalf("expected 3 ids, got %d", len(ids))
		}

		for i, id := range ids {
			searchFields := map[string]any{"id": id}
			iter, err := table.Search(&searchFields)
			if err != nil {
				t.Fatalf("Search failed for id %d: %v", id, err)
			}
			if iter == nil {
				t.Fatalf("Failed to get iterator for id %d", id)
			}
			defer GlobalTableIterPool.Put(iter)

			if !iter.Next() {
				t.Fatalf("record with id %d not found", id)
			}

			resultRecords := iter.GetRecords(false)
			if len(resultRecords) != 1 {
				t.Fatalf("expected 1 record for id %d, got %d", id, len(resultRecords))
			}

			result := resultRecords[0]
			expected := records[i]

			if result["name"] != (*expected)["name"] {
				t.Errorf("id %d: name mismatch in encrypted IoT DB, expected %v, got %v", id, (*expected)["name"], result["name"])
			}
			if result["age"] != (*expected)["age"] {
				t.Errorf("id %d: age mismatch in encrypted IoT DB, expected %v, got %v", id, (*expected)["age"], result["age"])
			}
			if result["email"] != (*expected)["email"] {
				t.Errorf("id %d: email mismatch in encrypted IoT DB, expected %v, got %v", id, (*expected)["email"], result["email"])
			}
		}
		t.Logf("Encrypted BatchInsertNoIncIoT test passed, all fields verified correctly")
	})

	t.Run("Encrypted_BatchInsertNoInc_ContinueStrategy", func(t *testing.T) {
		masterKey := make([]byte, 32)
		for i := range masterKey {
			masterKey[i] = byte(i)
		}

		encryptConfig := &storage.EncryptionConfig{
			Enabled:   true,
			Algorithm: "AES-256-GCM",
			MasterKey: masterKey,
		}

		dbPath := "./test_encrypted_batch_continue_db_" + t.Name()
		tableName := "test_encrypted_batch_continue_" + t.Name()

		_, err := storage.GetDBManager().OpenDB(dbPath, encryptConfig)
		if err != nil {
			t.Fatalf("Failed to open encrypted database: %v", err)
		}

		table, err := NewTable(tableName)
		if err != nil {
			t.Fatalf("NewTable failed: %v", err)
		}

		fields := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0}
		table.SetFields(fields)

		pk, _ := NewDefaultPrimaryKey("pk")
		pk.AddFields("id")
		table.CreateIndex(pk)

		records := []*map[string]any{
			{"id": 100, "name": "Valid1", "age": 20, "score": 95.5},
			{"name": "INVALID"},
			{"id": 101, "name": "Valid2", "age": 25, "score": 88.0},
			{"id": nil, "name": "INVALID"},
			{"id": 102, "name": "Valid3", "age": 30, "score": 91.25},
		}

		ids, err := table.BatchInsertNoInc(records, true)
		if err != nil {
			t.Fatalf("Encrypted Continue strategy failed: %v", err)
		}

		if len(ids) != 3 {
			t.Fatalf("expected 3 valid IDs, got %d: %v", len(ids), ids)
		}

		expectedRecords := []*map[string]any{
			{"id": 100, "name": "Valid1", "age": 20, "score": 95.5},
			{"id": 101, "name": "Valid2", "age": 25, "score": 88.0},
			{"id": 102, "name": "Valid3", "age": 30, "score": 91.25},
		}

		for i, id := range ids {
			searchFields := map[string]any{"id": id}
			iter, err := table.Search(&searchFields)
			if err != nil {
				t.Fatalf("Search failed for id %d: %v", id, err)
			}
			if iter == nil {
				t.Fatalf("Failed to get iterator for id %d", id)
			}
			defer GlobalTableIterPool.Put(iter)

			if !iter.Next() {
				t.Fatalf("record with id %d not found", id)
			}

			resultRecords := iter.GetRecords(false)
			if len(resultRecords) != 1 {
				t.Fatalf("expected 1 record for id %d, got %d", id, len(resultRecords))
			}

			result := resultRecords[0]
			expected := expectedRecords[i]

			if result["name"] != (*expected)["name"] {
				t.Errorf("id %d: name mismatch in encrypted DB, expected %v, got %v", id, (*expected)["name"], result["name"])
			}
			if result["age"] != (*expected)["age"] {
				t.Errorf("id %d: age mismatch in encrypted DB, expected %v, got %v", id, (*expected)["age"], result["age"])
			}
			if result["score"] != (*expected)["score"] {
				t.Errorf("id %d: score mismatch in encrypted DB, expected %v, got %v", id, (*expected)["score"], result["score"])
			}
		}
		t.Logf("Encrypted Continue strategy test passed, all valid records have correct data")
	})
}
