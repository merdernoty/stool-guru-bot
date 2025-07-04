package utils

import "fmt"

// Add adds two integers and returns the result.
func Add(a, b int) int {
    return a + b
}

// Subtract subtracts the second integer from the first and returns the result.
func Subtract(a, b int) int {
    return a - b
}

// Multiply multiplies two integers and returns the result.
func Multiply(a, b int) int {
    return a * b
}

// Divide divides the first integer by the second and returns the result.
// It returns an error if the second integer is zero.
func Divide(a, b int) (int, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}