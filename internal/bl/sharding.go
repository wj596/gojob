/*
 * Copyright 2020-2021 the original author(https://github.com/wj596)
 *
 * <p>
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
 * </p>
 */
package bl

// 分片
// 如果有3个实例 分为9片，每个实列得到的分片结果为：[[0 1 2] [3 4 5] [6 7 8]]
// 如果有3个实例 分为10片，每个实列得到的分片结果为：[[0 1 2 9] [3 4 5] [6 7 8]]
// 如果有3个实例 分为11片，每个实列得到的分片结果为：[[0 1 2 9] [3 4 5 10] [6 7 8]]
func Sharding(shardingTotal int, instanceCount int) [][]int {
	result := make([][]int, 0)
	entropy := shardingTotal / instanceCount
	index := 0
	for i := 0; i < instanceCount; i++ {
		items := make([]int, 0)
		for j := index * entropy; j < (index+1)*entropy; j++ {
			items = append(items, j)
		}
		result = append(result, items)
		index++
	}
	aliquant := shardingTotal % instanceCount
	index = 0
	for i := 0; i < instanceCount; i++ {
		if index < aliquant {
			v := shardingTotal/instanceCount*instanceCount + index
			result[i] = append(result[i], v)
		}
		index++
	}
	return result
}
