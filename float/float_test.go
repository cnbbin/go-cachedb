package float

import "testing"

func TestDecimalOperations(t *testing.T) {
	cases := []struct {
		a, b    string
		wantAdd string
		wantMul string
	}{
		{"1", "1", "2.0", "1.0"},
		{"3.14", "3.6666", "6.8066", "11.513124"},
		{"1234567890.123456", "0.000000001", "1234567890.123456001", "1.234567890123456"},
		{"-1.5", "2.5", "1.0", "-3.75"},
		{"1e3", "2e2", "1200.0", "200000.0"},
		{"123456789.123", "987654321.987", "1111111111.110", "121932631355968601.347401"}, // 答案待
		{"121932631355968601.123", "121932631355968601.123", "243865262711937202.246", "14867566589390537303347534211465476.861129"}, // 答案待
	}

	for _, c := range cases {
		d1, err := NewDecimal(c.a)
		if err != nil {
			t.Errorf("NewDecimal(%q) error: %v", c.a, err)
			continue
		}
		d2, err := NewDecimal(c.b)
		if err != nil {
			t.Errorf("NewDecimal(%q) error: %v", c.b, err)
			continue
		}

		addResult := d1.Add(d2).String()
		if addResult != c.wantAdd {
			t.Errorf("%q + %q = %q, want %q", c.a, c.b, addResult, c.wantAdd)
		}

		mulResult := d1.Mul(d2).String()
		if mulResult != c.wantMul {
			t.Errorf("%q * %q = %q, want %q", c.a, c.b, mulResult, c.wantMul)
		}
	}
}
