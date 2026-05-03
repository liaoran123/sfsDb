package engine

import (
	"testing"
)

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
		if err != nil {t.Fatalf("NewTable failed: %v", err)
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

		// 先测试 FailFast
		{
			_, testErr := table.BatchInsertInc(records, false)
			if testErr == nil {
				t.Fatal("expected error when invalid record with FailFast strategy, but got nil")
			}
			t.Logf("FailFast correctly returned error: %v", testErr)
		}
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
