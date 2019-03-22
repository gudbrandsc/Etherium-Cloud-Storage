package p1

import (
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
	"reflect"
	"strings"
)

type Flag_value struct {
	Encoded_prefix []uint8
	Value          string
}

type Node struct {
	Node_type    int // 0: Null, 1: Branch, 2: Ext or Leaf
	Branch_value [17]string
	Flag_value   Flag_value
}

type MerklePatriciaTrie struct {
	DB       map[string]Node
	Root     string
	EntryMap map[string]string
}

func (mpt *MerklePatriciaTrie) Get(key string) (string, error) {
	return mpt.get_node(mpt.Root, str_to_u8(key))
}

func (mpt *MerklePatriciaTrie) get_node(cur_hash string, path []uint8) (string, error) {
	node := mpt.DB[cur_hash]
	//fmt.Printf("cur_hash=%s, path=%v\n", cur_hash, path)
	switch node.Node_type {
	case 0:
		return "", errors.New("null")
	case 1:
		branch := node.Branch_value
		if len(path) == 0 {
			return branch[16], nil
		} else {
			first, rest := path[0], path[1:]
			next_hash := branch[first]
			if next_hash == "" {
				return "", errors.New("path_not_found")
			} else {
				return mpt.get_node(next_hash, rest)
			}
		}
	case 2:
		prefix := compact_decode(node.Flag_value.Encoded_prefix)
		//fmt.Printf("Flag Node: encoded_pre=%v, pre=%v, Value=%s\n", node.Flag_value.Encoded_prefix, prefix,
		//	node.Flag_value.Value)
		switch is_ext_node(node.Flag_value.Encoded_prefix) {
		case true:
			paths := split_path(path, prefix)
			rest_cur_path := paths[1]
			rest_path := paths[2]
			if len(rest_path) == 0 {
				return mpt.get_node(node.Flag_value.Value, rest_cur_path)
			} else {
				return "", errors.New("path_not_found")
			}
		case false:
			if reflect.DeepEqual(path, prefix) {
				return node.Flag_value.Value, nil
			} else {
				return "", errors.New("path_not_found")
			}
		}
	}
	return "error", errors.New("wrong_node_type")
}

func (mpt *MerklePatriciaTrie) Insert(key string, new_value string) {
	mpt.EntryMap[key] = new_value
	mpt.Root = mpt.insert_node(mpt.Root, str_to_u8(key), new_value)
}

func (mpt *MerklePatriciaTrie) insert_node(cur_hash string, path []uint8, new_value string) string {
	//fmt.Printf("path=%v, check node=%s\n", path, cur_hash)
	cur_node := mpt.DB[cur_hash]
	switch cur_node.Node_type {
	case 0:
		new_leaf := Node{Node_type: 2, Flag_value: Flag_value{Encoded_prefix: compact_encode(path_with_terminator(path)), Value: new_value}}
		new_hash := new_leaf.hash_node()
		//fmt.Printf("Convert Root::NULL to LEAF, new_hash=%s, new_node=%v\n", new_hash, new_leaf)
		delete(mpt.DB, cur_hash)
		mpt.DB[new_hash] = new_leaf
		return new_hash
	case 1:
		if len(path) == 0 || cur_node.Branch_value[path[0]] == "" {
			// update current BRANCH Value || add LEAF
			return mpt.make_leaf_from_branch(cur_hash, path, new_value)
		} else {
			rest_path := path[1:]
			next_hash := mpt.insert_node(cur_node.Branch_value[path[0]], rest_path, new_value)
			return mpt.update_branch_next_node(cur_hash, path[0], next_hash)
		}
	case 2:
		encoded_prefix := cur_node.Flag_value.Encoded_prefix
		ori_value := cur_node.Flag_value.Value
		prefix := compact_decode(encoded_prefix)
		//fmt.Printf("Flag Node: encoded_pre=%v, pre=%v, Value=%s\n", encoded_prefix, prefix, ori_value)
		switch is_ext_node(encoded_prefix) {
		case true:
			paths := split_path(prefix, path)
			common_path := paths[0]
			rest_cur_path := paths[1]
			rest_path := paths[2]
			if len(rest_cur_path) == 0 {
				next_hash := mpt.insert_node(ori_value, rest_path, new_value)
				return mpt.update_flag_node(cur_hash, prefix, next_hash)
			} else {
				return mpt.split_ext(cur_hash, common_path, rest_cur_path, rest_path, ori_value, new_value)
			}
		case false:
			if reflect.DeepEqual(path, prefix) {
				new_hash := mpt.make_leaf(path, new_value)
				delete(mpt.DB, cur_hash)
				//fmt.Printf("Update LEAF, old_hash=%s, new_hash=%s\n", cur_hash, new_hash)
				return new_hash
			} else {
				return mpt.split_leaf(cur_hash, prefix, ori_value, path, new_value)
			}
		}
	}
	return "error"
}

func (mpt *MerklePatriciaTrie) update_flag_node(cur_hash string, path []uint8, value string) string {
	new_node := Node{Node_type: 2, Flag_value: Flag_value{Encoded_prefix: compact_encode(path), Value: value}}
	new_hash := new_node.hash_node()
	//fmt.Printf("Update FLAG, old_hash=%s, new_hash=%s, new_node=%s\n", cur_hash, new_hash, node_to_string(new_node))
	delete(mpt.DB, cur_hash)
	mpt.DB[new_hash] = new_node
	return new_hash
}

// From a BRANCH, add LEAF or update Value, return new_branch_hash
func (mpt *MerklePatriciaTrie) make_leaf_from_branch(cur_hash string, path []uint8, new_value string) string {
	if len(path) == 0 {
		return mpt.update_branch_value(cur_hash, new_value)
	} else {
		new_leaf_path := path[1:]
		new_hash := mpt.make_leaf(new_leaf_path, new_value)
		return mpt.update_branch_next_node(cur_hash, path[0], new_hash)
	}
}

func (mpt *MerklePatriciaTrie) make_leaf(path []uint8, new_value string) string {
	new_leaf := Node{Node_type: 2, Flag_value: Flag_value{Encoded_prefix: compact_encode(path_with_terminator(path)), Value: new_value}}
	new_hash := new_leaf.hash_node()
	//fmt.Printf("Insert LEAF, new_hash=%s, new_node=%v\n", new_hash, new_leaf)
	mpt.DB[new_hash] = new_leaf
	return new_hash
}

func (mpt *MerklePatriciaTrie) make_ext(path []uint8, new_value string) string {
	new_leaf := Node{Node_type: 2, Flag_value: Flag_value{Encoded_prefix: compact_encode(path), Value: new_value}}
	new_hash := new_leaf.hash_node()
	//fmt.Printf("Insert Ext, new_hash=%s, new_node=%v\n", new_hash, new_leaf);
	mpt.DB[new_hash] = new_leaf
	return new_hash
}

func (mpt *MerklePatriciaTrie) update_branch_value(cur_hash string, new_value string) string {
	node := mpt.DB[cur_hash]
	node.Branch_value[16] = new_value
	new_hash := node.hash_node()
	//fmt.Printf("Update BRANCH, old_hash=%s, new_hash=%s, new_node=%s\n", cur_hash, new_hash, node_to_string(node))
	delete(mpt.DB, cur_hash)
	mpt.DB[new_hash] = node
	return new_hash
}

func (mpt *MerklePatriciaTrie) update_branch_next_node(cur_hash string, new_index uint8, new_next_hash string) string {
	node := mpt.DB[cur_hash]
	node.Branch_value[new_index] = new_next_hash
	new_hash := node.hash_node()
	//fmt.Printf("Update BRANCH, old_hash=%s, new_hash=%s, new_node=%s\n", cur_hash, new_hash, node_to_string(node))
	delete(mpt.DB, cur_hash)
	mpt.DB[new_hash] = node
	return new_hash
}

// Update old LEAF to EXT+BRANCH, add old_value and new_value, return the most top node's hash
func (mpt *MerklePatriciaTrie) split_leaf(cur_hash string, leaf_prefix []uint8, leaf_value string, path []uint8,
	new_value string) string {
	//fmt.Println("Start splitting LEAF(" + cur_hash + ")")
	paths := split_path(leaf_prefix, path)
	common_path := paths[0]
	rest_cur_path := paths[1]
	rest_path := paths[2]
	delete(mpt.DB, cur_hash)

	new_branch := Node{Node_type: 1, Branch_value: empty_branch_value()}
	new_branch_hash := new_branch.hash_node()
	mpt.DB[new_branch_hash] = new_branch
	new_branch_hash = mpt.make_leaf_from_branch(new_branch_hash, rest_cur_path, leaf_value)
	new_branch_hash = mpt.make_leaf_from_branch(new_branch_hash, rest_path, new_value)

	if len(common_path) == 0 {
		// branch
		//fmt.Printf("Split LEAF to BRANCH, old_hash=%s, new_hash=%s\n", cur_hash, new_branch_hash)
		return new_branch_hash
	} else {
		// ext+branch
		new_ext := Node{Node_type: 2, Flag_value: Flag_value{Encoded_prefix: compact_encode(common_path), Value: new_branch_hash}}
		new_ext_hash := new_ext.hash_node()
		mpt.DB[new_ext_hash] = new_ext
		//fmt.Printf("Split LEAF(%s) to EXT(%s) + BRANCH(%s)\n", cur_hash, new_ext_hash, new_branch_hash)
		return new_ext_hash
	}
	return "error"
}

// Update old EXT -> NEXT_BRANCH to EXT + BRANCH(LEAF + EXT->NEXT_BRANCH)
// add old_value and new_value, return the most top node's hash
func (mpt *MerklePatriciaTrie) split_ext(cur_hash string, common_path []uint8, rest_cur_path []uint8, rest_path []uint8,
	ori_value string, new_value string) string {
	delete(mpt.DB, cur_hash)

	new_branch := Node{Node_type: 1, Branch_value: empty_branch_value()}
	new_branch_hash := new_branch.hash_node()
	mpt.DB[new_branch_hash] = new_branch

	// append EXT->NEXT_BRANCH to BRANCH, then append LEAF
	if len(rest_cur_path) == 1 {
		// NEXT_BRANCH
		new_branch_hash = mpt.update_branch_next_node(new_branch_hash, rest_cur_path[0], ori_value)
	} else {
		// EXT->NEXT_BRANCH
		new_ext_path := rest_cur_path[1:]
		new_ext_hash := mpt.update_flag_node(cur_hash, new_ext_path, ori_value)
		new_branch_hash = mpt.update_branch_next_node(new_branch_hash, rest_cur_path[0], new_ext_hash)
	}
	new_branch_hash = mpt.make_leaf_from_branch(new_branch_hash, rest_path, new_value)

	if len(common_path) == 0 {
		// BRANCH
		//fmt.Printf("Split EXT(%s) to BRANCH(%s)", cur_hash, new_branch_hash)
		return new_branch_hash
	} else {
		// EXT-+BRANCH
		new_ext := Node{Node_type: 2, Flag_value: Flag_value{Encoded_prefix: compact_encode(common_path), Value: new_branch_hash}}
		new_ext_hash := new_ext.hash_node()
		//fmt.Printf("Split EXT(%s) to EXT(%s) + BRANCH(%s)", cur_hash, new_ext_hash, new_branch_hash)
		mpt.DB[new_ext_hash] = new_ext
		return new_ext_hash
	}
}

func (mpt *MerklePatriciaTrie) check_branch_only_one_value(cur_hash string) (bool, int32) {
	cur_node := mpt.DB[cur_hash]
	count := 0
	first_index := 20
	for i := 0; i < 17; i++ {
		if cur_node.Branch_value[i] != "" {
			count += 1
			first_index = i
		}
	}
	//fmt.Printf("Check Branch(%s) only one Value: %v, first_index=%v\n", cur_hash, count <= 1, first_index)
	return count <= 1, int32(first_index)
}

func (mpt *MerklePatriciaTrie) Delete(key string) string {
	delete(mpt.EntryMap, key)
	msg := mpt.delete_node(mpt.Root, str_to_u8(key))
	if msg != "path_not_found" {
		mpt.Root = msg
		//fmt.Println("Update ROOT to " + mpt.Root)
		return ""
	} else {
		return "path_not_found"
	}
}

func (mpt *MerklePatriciaTrie) delete_node(cur_hash string, path []uint8) string {
	//fmt.Printf("delete_node(), cur_hash=%s, path=%v\n", cur_hash, path)
	cur_node := mpt.DB[cur_hash]
	switch cur_node.Node_type {
	case 0:
		return "path_not_found"
	case 1:
		branch := cur_node.Branch_value
		if len(path) == 0 {
			new_cur_hash := mpt.update_branch_value(cur_hash, "")
			if_branch_one_value, branch_first_index := mpt.check_branch_only_one_value(new_cur_hash)
			if if_branch_one_value {
				return mpt.merge_branch_down(new_cur_hash, branch_first_index)
			} else {
				return new_cur_hash
			}
		} else {
			rest := path[1:]
			if branch[path[0]] == "" {
				return "path_not_found"
			}
			new_next_hash := mpt.delete_node(branch[path[0]], rest)
			if new_next_hash == "" {
				new_cur_hash := mpt.update_branch_next_node(cur_hash, path[0], new_next_hash)
				if_branch_one_value, branch_first_index := mpt.check_branch_only_one_value(new_cur_hash)
				if if_branch_one_value {
					return mpt.merge_branch_down(new_cur_hash, branch_first_index)
				} else {
					return new_cur_hash
				}
			} else if new_next_hash == "path_not_found" {
				return "path_not_found"
			} else {
				return mpt.update_branch_next_node(cur_hash, path[0], new_next_hash)
			}
		}
	case 2:
		encoded_prefix := cur_node.Flag_value.Encoded_prefix
		ori_value := cur_node.Flag_value.Value
		prefix := compact_decode(encoded_prefix)
		//fmt.Printf("Flag Node: encoded_pre=%v, pre=%v, Value=%s\n", encoded_prefix, prefix, ori_value)
		switch is_ext_node(encoded_prefix) {
		case true:
			paths := split_path(prefix, path)
			rest_cur_path := paths[1]
			rest_path := paths[2]
			if len(rest_cur_path) != 0 {
				return "path_not_found"
			} else {
				new_next_hash := mpt.delete_node(ori_value, rest_path)
				if new_next_hash == "" {
					delete(mpt.DB, cur_hash)
					return ""
				} else if new_next_hash == "path_not_found" {
					return "path_not_found"
				} else {
					new_cur_hash := mpt.update_flag_node(cur_hash, prefix, new_next_hash)
					next_node := mpt.DB[new_next_hash]
					switch next_node.Node_type {
					case 1:
						return new_cur_hash
					case 2:
						//fmt.Println("Cur is Ext, next is Flag, merge ext down")
						return mpt.merge_ext_down(new_cur_hash, new_next_hash)
					}
				}
			}
		case false:
			if reflect.DeepEqual(prefix, path) {
				delete(mpt.DB, cur_hash)
				return ""
			} else {
				return "path_not_found"
			}
		}
	}
	return ""
}

// For Delete(), if Branch only one Value then merge down, return new_node_hash
// branch+ext => ext
// branch+leaf => leaf
// branch+branch => ext+branch
// branch => leaf
func (mpt *MerklePatriciaTrie) merge_branch_down(cur_hash string, first_index int32) string {
	cur_node := mpt.DB[cur_hash]
	branch := cur_node.Branch_value
	new_hash := ""
	if first_index == 16 {
		//fmt.Println("merge_branch_down() convert BRANCH to LEAF")
		new_hash = mpt.make_leaf([]uint8{}, branch[16])
	} else {
		next_node := mpt.DB[branch[first_index]]
		switch next_node.Node_type {
		case 1:
			//fmt.Println("merge_branch_down(), next is BRANCH")
			new_hash = mpt.make_ext([]uint8{uint8(first_index)}, branch[first_index])
		case 2:
			encoded_prefix := next_node.Flag_value.Encoded_prefix
			ori_value := next_node.Flag_value.Value
			new_path := arr_add([]uint8{uint8(first_index)}, compact_decode(encoded_prefix))
			if !is_ext_node(encoded_prefix) {
				new_path = append(new_path, 16)
				//fmt.Printf("Next node is a LEAF, new_path=%v\n", new_path)
			}
			new_hash = mpt.update_flag_node(branch[first_index], new_path, ori_value)
		}
	}
	delete(mpt.DB, cur_hash)
	return new_hash
}

// ext+leaf => leaf
// ext+ext => lower_ext
func (mpt *MerklePatriciaTrie) merge_ext_down(cur_hash string, next_hash string) string {
	cur_node := mpt.DB[cur_hash]
	next_node := mpt.DB[next_hash]
	encoded_prefix := cur_node.Flag_value.Encoded_prefix
	new_path := compact_decode(encoded_prefix)

	delete(mpt.DB, cur_hash)

	encoded_prefix = next_node.Flag_value.Encoded_prefix
	ori_value := next_node.Flag_value.Value
	prefix := compact_decode(encoded_prefix)
	new_path = arr_add(new_path, prefix)
	if !is_ext_node(encoded_prefix) {
		new_path = append(new_path, 16)
	}
	//fmt.Printf("Merge_ext_down(), path=%v\n", new_path)
	return mpt.update_flag_node(next_hash, new_path, ori_value)
}

func compact_encode(hex_array []uint8) []uint8 {
	//fmt.Printf("Encode %v", hex_array)
	term := 0
	if hex_array[len(hex_array)-1] == 16 {
		term = 1
		hex_array = hex_array[:len(hex_array)-1]
	}
	odd_len := len(hex_array) % 2
	flags := uint8(2*term + odd_len)
	if odd_len == 1 {
		hex_array = arr_add([]uint8{flags}, hex_array)
	} else {
		hex_array = arr_add([]uint8{flags, 0}, hex_array)
	}
	rs := []uint8{}
	i := 0
	for i < len(hex_array) {
		rs = append(rs, hex_array[i]*16+hex_array[i+1])
		i += 2
	}
	//fmt.Printf(" to %v, ASCII=%v\n", hex_array, rs)
	return rs
}

func compact_decode(encoded_arr []uint8) []uint8 {
	decoded_arr := []uint8{}
	for _, each := range encoded_arr {
		decoded_arr = append(decoded_arr, each/16)
		decoded_arr = append(decoded_arr, each%16)
	}
	if decoded_arr[0]%2 == 0 {
		return decoded_arr[2:]
	} else {
		return decoded_arr[1:]
	}
}

func str_to_u8(input string) []uint8 {
	rs := []uint8{}
	for _, each := range input {
		rs = append(rs, uint8(int(each)/16))
		rs = append(rs, uint8(int(each)%16))
	}
	return rs
}

func path_with_terminator(path []uint8) []uint8 {
	return append(path, 16)
}

func TestCompact2() {
	test_compact_encode()
	test_split_path()
}

func test_compact_encode() {
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{1, 2, 3, 4, 5})), []uint8{1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 1, 2, 3, 4, 5})), []uint8{0, 1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 15, 1, 12, 11, 8, 16})), []uint8{0, 15, 1, 12, 11, 8}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{15, 1, 12, 11, 8, 16})), []uint8{15, 1, 12, 11, 8}))
	fmt.Println(reflect.DeepEqual(compact_encode([]uint8{6, 1}), []uint8{0, 97}))
	fmt.Println(reflect.DeepEqual(compact_encode([]uint8{2, 16}), []uint8{50}))
}

func test_split_path() {
	fmt.Println(reflect.DeepEqual(split_path([]uint8{1, 2, 3}, []uint8{1, 2, 3}), [3][]uint8{{1, 2, 3}, {}, {}}))
	fmt.Println(reflect.DeepEqual(split_path([]uint8{1, 2, 3, 4}, []uint8{1, 2, 3}), [3][]uint8{{1, 2, 3}, {4}, {}}))
	fmt.Println(reflect.DeepEqual(split_path([]uint8{1, 2, 3}, []uint8{1, 2, 3, 4}), [3][]uint8{{1, 2, 3}, {}, {4}}))
	fmt.Println(reflect.DeepEqual(split_path([]uint8{1, 2, 3, 4}, []uint8{2, 3, 4, 5}), [3][]uint8{{}, {1, 2, 3, 4}, {2, 3, 4, 5}}))
	fmt.Println(reflect.DeepEqual(split_path([]uint8{1, 2, 3, 4}, []uint8{1, 2, 4, 5}), [3][]uint8{{1, 2}, {3, 4}, {4, 5}}))
}

func split_path(path1 []uint8, path2 []uint8) [3][]uint8 {
	i := 0
	for i < len(path1) && i < len(path2) {
		if path1[i] == path2[i] {
			i += 1
		} else {
			break
		}
	}
	common := path1[:i]
	rest1 := path1[i:]
	rest2 := path2[i:]
	//fmt.Printf("Split path (%v, %v) to common=%v, rest_cur_path=%v, rest_path=%v\n", path1, path2, common, rest1, rest2)
	return [3][]uint8{common, rest1, rest2}
}

func (node *Node) hash_node() string {
	var str string
	switch node.Node_type {
	case 0:
		str = ""
	case 1:
		str = "branch_"
		for _, v := range node.Branch_value {
			str += v
		}
	case 2:
		str = node.Flag_value.Value
	}

	sum := sha3.Sum256([]byte(str))
	return "HashStart_" + hex.EncodeToString(sum[:]) + "_HashEnd"
}

func empty_branch_value() [17]string {
	return [17]string{"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""}
}

func arr_add(arr1 []uint8, arr2 []uint8) []uint8 {
	for _, each := range arr2 {
		arr1 = append(arr1, each)
	}
	return arr1
}

func (node *Node) String() string {
	str := "empty string"
	switch node.Node_type {
	case 0:
		str = "[Null Node]"
	case 1:
		str = "Branch["
		for i, v := range node.Branch_value[:16] {
			str += fmt.Sprintf("%d=\"%s\", ", i, v)
		}
		str += fmt.Sprintf("Value=%s]", node.Branch_value[16])
	case 2:
		encoded_prefix := node.Flag_value.Encoded_prefix
		node_name := "Leaf"
		if is_ext_node(encoded_prefix) {
			node_name = "Ext"
		}
		ori_prefix := strings.Replace(fmt.Sprint(compact_decode(encoded_prefix)), " ", ", ", -1)
		str = fmt.Sprintf("%s<%v, Value=\"%s\">", node_name, ori_prefix, node.Flag_value.Value)
	}
	return str
}

func node_to_string(node Node) string {
	return node.String()
}

func (mpt *MerklePatriciaTrie) Initial() {
	mpt.DB = make(map[string]Node)
	mpt.Root = ""
	mpt.EntryMap = make(map[string]string)
}

func is_ext_node(encoded_arr []uint8) bool {
	return encoded_arr[0]/16 < 2
}

func TestCompact() {
	test_compact_encode()
}

func (mpt *MerklePatriciaTrie) String() string {
	content := fmt.Sprintf("ROOT=%s\n", mpt.Root)
	for hash := range mpt.DB {
		content += fmt.Sprintf("%s: %s\n", hash, node_to_string(mpt.DB[hash]))
	}
	return content
}

func (mpt *MerklePatriciaTrie) Order_nodes() string {
	raw_content := mpt.String()
	content := strings.Split(raw_content, "\n")
	root_hash := strings.Split(strings.Split(content[0], "HashStart")[1], "HashEnd")[0]
	queue := []string{root_hash}
	i := -1
	rs := ""
	cur_hash := ""
	for len(queue) != 0 {
		last_index := len(queue) - 1
		cur_hash, queue = queue[last_index], queue[:last_index]
		i += 1
		line := ""
		for _, each := range content {
			if strings.HasPrefix(each, "HashStart"+cur_hash+"HashEnd") {
				line = strings.Split(each, "HashEnd: ")[1]
				rs += each + "\n"
				rs = strings.Replace(rs, "HashStart"+cur_hash+"HashEnd", fmt.Sprintf("Hash%v", i), -1)
			}
		}
		temp2 := strings.Split(line, "HashStart")
		flag := true
		for _, each := range temp2 {
			if flag {
				flag = false
				continue
			}
			queue = append(queue, strings.Split(each, "HashEnd")[0])
		}
	}
	rs = strings.Replace(rs, "\r\n", "\n", -1)
	return rs
}

func (mpt *MerklePatriciaTrie) GetEntryMap() map[string]string {
	return mpt.EntryMap
}
