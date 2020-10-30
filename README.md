# sfpc
sfpc is a lossy *simple floating point compressor* for individual float numbers.
Especially when working with decimals in sql-based contexts, there are often
values like 0.2 which cannot be represented in an accurate way using a float. Instead,
the exact decimal, like 0.2 is represented as 0.200000000000000011102230246251565404236316680908203125. 
However, IEEE 754 float guarantees that un-presentable numbers are "printed" correctly.

The conclusion is, that if we know, that a float represents such a decimal, we can compress it correctly and 
efficiently. Any additional precision is nothing but noise, which has been artificially 
introduced and can be removed safely.

The algorithm works as follows:
* there are 4 levels of accuracy: 10^
* values between -128 and 126 without a fraction within the given decimal accuracy are encoded directly as is
* any other values are encoded as a scaled signed integer in varint zigzag encoding
* if the float exceeds the range of the decimal encoding or the representation would exceed 8 bytes, the
float is just stored. 
