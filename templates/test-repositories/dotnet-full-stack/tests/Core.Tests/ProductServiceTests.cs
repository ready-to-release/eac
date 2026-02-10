using Core.Services;
using Xunit;

namespace Core.Tests;

public class ProductServiceTests
{
    [Fact]
    public void Add_ReturnsProduct()
    {
        var service = new ProductService();
        var product = service.Add("Widget", 9.99m);

        Assert.NotNull(product);
        Assert.Equal("Widget", product.Name);
        Assert.Equal(9.99m, product.Price);
    }

    [Fact]
    public void GetAll_ReturnsAddedProducts()
    {
        var service = new ProductService();
        service.Add("Widget", 9.99m);
        service.Add("Gadget", 19.99m);

        Assert.Equal(2, service.GetAll().Count);
    }
}
