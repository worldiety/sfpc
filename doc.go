/*
 * Copyright 2020 Torben Schinke
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package sfpc - *simple floating point compressor* - prodivdes a lossy variable length encoding
// for individual double precision float numbers with a fractional accuracy of 10^-9.
//
// motivation
//
// When dealing with sensor metrics, like SCADA systems, numbers are often represented
// using fixed point decimal metrics. However, in data analytics, these metrics are
// usually converted and processed in double precision floating-point values. Even
// though representing decimals using floating-point values is highly inaccurate and
// in general not recommended, it is just practical. Consider the value 0.2 which
// is not possible to represent as a float. Instead, even a float64 can only store
// a similar value with a lot of noise
// (actually 0.200000000000000011102230246251565404236316680908203125).
// However, an IEEE 754 float guarantees that un-presentable numbers are "printed" correctly.
//
// We evaluated, that in our use cases, 80% of our data is located in the decimal
// range of 0.000 up to 100.000.
//
// concept
//
// We learned, that a float contains a lot of noise, even for very small numbers near zero.
// We know, that noise is bad for compression. But, we also learned, that the noise
// is irrelevant at the first place - until it starts accumulating, which is not our subject here.
// Taking for example all natural numbers, we can just apply a variable length encoding
// for integer, which is a well known and solved problem space. If we know the scaling
// of a float number like 0.2 we can apply the same idea, just by upscaling by 10x and
// applying a variable integer encoding as well. The disadvantage is, that we need a prefix
// to indicate the kind of encoding. Using a naive encoding, we can represent 0.2 exactly
// using a prefix byte and a value byte, which is not even possible for
// a double precision 8 byte floating-point which fills everything else with
// garbage noise. Our encoding just takes 1/4 of the space. This sounds not much, but
// storing 1 billion 8 byte floats would require 7.45GiB of memory but if representable
// in the two byte range this would only require 1.86GiB of memory instead and you
// can still be sure, that the precision of the fraction is not worse than 10^-9.
//
// encoding characteristics
//
// As a general rule, we always store the original float64, if we cannot guarantee a
// fractional accuracy of 10^-9, which is the lossless part of our procedure. This is
// also the worst case, where we need 9 byte instead of 8, due to our prefix encoding.
// However, in the following cases, we need less bytes:
//
//  * any natural number between -114 and 128 is encoded directly in the prefix byte.
//  * special values like +Inf, -Inf and NaN are encoded directly in the prefix byte.
//  * any larger natural number is encoded using the protobuf varuint encoding, but not using the
//    signed zigzag encoding. We sacrifice the bits from our prefix byte, to allow a
//    better space utilization in our expected number range. Encodings which can be represented
//    as a float32 and take less space, are favored. Encodings which needs more than 8
//    byte, are discarded and the original float64 value is stored.
//  * decimals with fractions up to 10^4 are scaled and rounded, if their accuracy is
//    still smaller than 10^-9. Afterwards, also the varuint encoding is applied.
package sfpc
