# stock-exchange

This is a Golang implementation of a limit order book, which is a data structure that keeps track of buy and sell orders for a financial asset. A limit order book consists of two sides: the bid side and the ask side. The bid side contains orders to buy the asset at a specified price or lower, while the ask side contains orders to sell the asset at a specified price or higher. The orders are sorted by price and time priority, meaning that the best available price is always at the top of the book, and the oldest order at the same price level has the highest priority.

The code defines several types and methods to represent and manipulate the limit order book. The main types are:

Match: a struct that represents a trade between a buy and a sell order, with the size, price, and the orders involved.
Order: a struct that represents an order to buy or sell a certain size of the asset, with a timestamp, a boolean flag indicating whether it is a bid or an ask, and a pointer to the limit level it belongs to.
Limit: a struct that represents a price level in the limit order book, with the price, a slice of orders, and the total volume at that level.
Orderbook: a struct that represents the limit order book itself, with two slices of limits for the bid and ask sides, and two maps that store the limits by price for fast lookup.
The code also defines some helper types and methods to sort the orders and limits by price and time priority, such as Orders, Limits, ByBestAsk, ByBestBid, and their corresponding Len, Swap, and Less methods.

The main methods of the Orderbook type are:

PlaceOrder: a method that takes a price and an order as arguments, and tries to match the order with the opposite side of the book. If the order is fully or partially filled, it returns a slice of matches. If the order is not fully filled, it adds the remaining size to the appropriate limit level in the book.
add: a helper method that takes a price and an order as arguments, and adds the order to the corresponding limit level in the book. If the limit level does not exist, it creates a new one and updates the slices and maps accordingly.
The code also defines some methods for the Order and Limit types, such as:

NewOrder: a function that creates a new order with the given size, bid flag, and current timestamp.
String: a method that returns a string representation of an order, showing its size.
NewLimit: a function that creates a new limit with the given price and an empty slice of orders.
AddOrder: a method that adds an order to a limit, updating its pointer, slice, and total volume.
DeleteOrder: a method that deletes an order from a limit, removing its pointer, slice, and total volume, and resorting the slice by time priority.


A limit order book is a record of outstanding orders to buy or sell a financial asset at a specific price or better. A limit order book consists of two sides: the bid side and the ask side. The bid side contains orders to buy the asset at a specified price or lower, while the ask side contains orders to sell the asset at a specified price or higher. The orders are sorted by price and time priority, meaning that the best available price is always at the top of the book, and the oldest order at the same price level has the highest priority. A limit order book is maintained by a specialist or an electronic system that matches and executes the orders according to the price and time rules. A limit order book helps traders to control the prices at which they trade and to see the depth and liquidity of the market.

The sort.Sort(Orders) function is a generic function that sorts a collection of elements that implements the sort.Interface interface. The sort.Interface interface requires three methods: Len, Less and Swap, which define the length, the comparison and the swapping of elements in the collection. The Orders type is a slice of pointers to Order structs, and it implements the sort.Interface interface by defining the Len, Less and Swap methods for Orders. The Len method returns the number of elements in the Orders slice, the Less method compares the Timestamp fields of two elements and returns true if the first element is older than the second element, and the Swap method exchanges the positions of two elements in the Orders slice. The sort.Sort(Orders) function uses the quicksort algorithm to sort the Orders slice in increasing order of Timestamp, which means that the oldest order has the highest priority. The quicksort algorithm is a recursive algorithm that divides the collection into two subcollections based on a pivot element, such that all elements in the left subcollection are less than or equal to the pivot element, and all elements in the right subcollection are greater than or equal to the pivot element. Then, the algorithm sorts the left and right subcollections recursively until the collection is sorted.

The PlaceOrder function in Orderbook struct is a function that takes a price and an order as parameters and tries to match the order with the existing orders in the order book. The function has two steps:

The first step is to try to match the order with the opposite side of the order book. For example, if the order is a buy order, the function will try to match it with the sell orders in the order book. The matching logic is based on the price and the priority of the orders. The function will try to find the best price for the order, which is the lowest price for a buy order and the highest price for a sell order. If there are multiple orders at the same price level, the function will use the timestamp of the orders to determine the priority, and match the order with the oldest order first. The function will continue to match the order until either the order is fully filled or there are no more matching orders in the order book. The function will return a slice of Match structs, which contain the information about the matched orders, such as the ask order, the bid order, the size filled and the price.

The second step is to add the rest of the order to the order book, if the order is not fully filled in the first step. The function will check if there is an existing limit level for the order’s price in the order book. If there is, the function will add the order to the limit level’s orders slice and update the limit level’s total volume. If there is not, the function will create a new limit level for the order’s price and add the order to the limit level’s orders slice. The function will also update the order book’s asks or bids slice and the ask limits or bid limits map, depending on whether the order is a buy or sell order. The function will use the sort package to sort the asks or bids slice in ascending or descending order of price, respectively.

A Limit Order is a  way to provide liquidity to the exchange