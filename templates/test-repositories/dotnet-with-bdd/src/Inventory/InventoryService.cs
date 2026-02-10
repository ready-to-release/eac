namespace Inventory;

public class InventoryService
{
    private readonly Dictionary<string, int> _stock = new();

    public int GetStock(string item)
    {
        return _stock.TryGetValue(item, out var count) ? count : 0;
    }

    public void AddStock(string item, int quantity)
    {
        if (!_stock.ContainsKey(item))
            _stock[item] = 0;
        _stock[item] += quantity;
    }

    public void RemoveStock(string item, int quantity)
    {
        var current = GetStock(item);
        if (quantity > current)
            throw new InvalidOperationException("Insufficient stock");
        _stock[item] = current - quantity;
    }
}
