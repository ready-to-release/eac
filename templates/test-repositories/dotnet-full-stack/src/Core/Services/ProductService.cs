namespace Core.Services;

using Core.Models;

public class ProductService
{
    private readonly List<Product> _products = new();

    public IReadOnlyList<Product> GetAll() => _products.AsReadOnly();

    public Product? GetById(string id) => _products.FirstOrDefault(p => p.Id == id);

    public Product Add(string name, decimal price)
    {
        var product = new Product(Guid.NewGuid().ToString(), name, price);
        _products.Add(product);
        return product;
    }
}
