using MyLibrary;
using Xunit;

namespace MyLibrary.Tests;

public class CalculatorTests
{
    private readonly Calculator _calculator = new();

    [Fact]
    public void Add_ReturnsCorrectSum()
    {
        Assert.Equal(5, _calculator.Add(2, 3));
    }

    [Fact]
    public void Subtract_ReturnsCorrectDifference()
    {
        Assert.Equal(1, _calculator.Subtract(3, 2));
    }

    [Fact]
    public void Divide_ByZero_ThrowsException()
    {
        Assert.Throws<DivideByZeroException>(() => _calculator.Divide(1, 0));
    }

    [Theory]
    [InlineData(10, 2, 5.0)]
    [InlineData(7, 3, 2.333333)]
    public void Divide_ReturnsCorrectQuotient(int a, int b, double expected)
    {
        Assert.Equal(expected, _calculator.Divide(a, b), 5);
    }
}
