package models

import "testing"

func TestStackOutput_IsCompatibleWithKubeSecret(t *testing.T) {
	tests := []struct {
		name   string
		output StackOutput
		want   bool
	}{
		{
			name: "valid",
			output: StackOutput{
				Id: "foobar",
			},
			want: true,
		},
		{
			name: "valid-alphanum",
			output: StackOutput{
				Id: "Foobar123",
			},
			want: true,
		},
		{
			name: "valid-dash",
			output: StackOutput{
				Id: "foobar-",
			},
			want: true,
		},
		{
			name: "valid underscore",
			output: StackOutput{
				Id: "foobar_",
			},
			want: true,
		},
		{
			name: "valid dot",
			output: StackOutput{
				Id: "foobar.",
			},
			want: true,
		},

		// Invalid ones
		{
			name: "invalid empty",
			output: StackOutput{
				Id: "",
			},
		},
		{
			name: "invalid special char",
			output: StackOutput{
				Id: "foobar!",
			},
		},
		{
			name: "invalid space",
			output: StackOutput{
				Id: " foobar",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.output.IsCompatibleWithKubeSecret(); got != tt.want {
				t.Errorf("IsCompatibleWithKubeSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}
