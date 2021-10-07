package formulate

import "testing"

func TestMemoryValidationStore_AddValidationError(t *testing.T) {
	v := NewMemoryValidationStore()

	err := v.AddValidationError("AddressLine1", ValidationError{Error: "Invalid address", Value: "1 Fake Street"})

	if err != nil {
		t.Error(err)
		return
	}

	err = v.AddValidationError("AddressLine2", ValidationError{Error: "Value must not be empty", Value: ""})

	if err != nil {
		t.Error(err)
		return
	}

	err = v.AddValidationError("AddressLine1", ValidationError{Error: "Fake Street not found", Value: "1 Fake Street"})

	if err != nil {
		t.Error(err)
		return
	}

	validationErrors, err := v.GetValidationErrors("AddressLine1")

	if err != nil {
		t.Error(err)
		return
	}

	if len(validationErrors) != 2 {
		t.Fail()
	}

	assertEquals(t, validationErrors[0].Error, "Invalid address")
	assertEquals(t, validationErrors[1].Error, "Fake Street not found")

	validationErrors, err = v.GetValidationErrors("AddressLine2")

	if err != nil {
		t.Error(err)
		return
	}

	if len(validationErrors) != 1 {
		t.Fail()
	}

	assertEquals(t, validationErrors[0].Error, "Value must not be empty")

	validationErrors, err = v.GetValidationErrors("NotFound")

	assertEquals(t, len(validationErrors), 0)
}

func TestMemoryValidationStore_GetFormValue(t *testing.T) {
	t.Run("Saved value is pointer", func(t *testing.T) {
		v := NewMemoryValidationStore()

		value := &YourDetails{Name: "Test"}

		if err := v.SetFormValue(value); err != nil {
			t.Error(err)
			return
		}

		var value2 YourDetails

		if err := v.GetFormValue(&value2); err != nil {
			t.Error(err)
			return
		}

		assertEquals(t, value2.Name, "Test")
	})

	t.Run("Saved value is not pointer", func(t *testing.T) {
		v := NewMemoryValidationStore()

		value := YourDetails{Name: "Test2"}

		if err := v.SetFormValue(value); err != nil {
			t.Error(err)
			return
		}

		var value2 YourDetails

		if err := v.GetFormValue(&value2); err != nil {
			t.Error(err)
			return
		}

		assertEquals(t, value2.Name, "Test2")
	})

	t.Run("Get value is not pointer", func(t *testing.T) {
		v := NewMemoryValidationStore()

		value := YourDetails{Name: "Test2"}

		if err := v.SetFormValue(value); err != nil {
			t.Error(err)
			return
		}

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic() on non-ptr type.")
			}
		}()

		var value2 YourDetails

		if err := v.GetFormValue(value2); err != nil {
			t.Error(err)
			return
		}
	})
}
