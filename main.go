package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"golang.org/x/text/unicode/norm"
	"hash/fnv"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	filep   string
	savep   string
	similar float64
	wt      string
)

const epsilon = 1e-7

func main() {
	flag.StringVar(&filep, "f", "./urls.txt", "指定待去重url文件路径")
	flag.StringVar(&savep, "o", "./output.txt", "去重结果输出至文件")
	flag.Float64Var(&similar, "s", 0.95, "指定相似度,去除比较结果中高于该相似度的url")
	flag.StringVar(&wt, "p", "4:3:2:0.5:0.5", "自定义host:path:param:frag:scheme的比例,请参照默认值格式输入")
	flag.Parse()
	//权重改为列表
	wa := strings.Split(wt, ":")
	if len(wa) == 5 {
		// 检查文件是否存在
		if _, err := os.Stat(filep); err == nil {
			url_de(filep, savep, similar, wa)
		} else if os.IsNotExist(err) {
			fmt.Println("文件不存在!")
		} else {
			fmt.Println("发生错误:", err)
		}
	} else {
		fmt.Println("请输入正确格式的数据")
	}

}
func url_de(filep string, savep string, similar float64, weight_arr []string) {
	//文件路径
	filePath := filep
	//文件保存地址
	savePath := savep

	//文件里读出url
	urls, err := inputFile(filePath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	//遍历url 得到Simhash结果
	hashes := make([]uint64, len(urls))
	for i, single_url := range urls {
		hashes[i] = Fingerprint(Vectorize(GetFeaturesFromURI(string(single_url), weight_arr)))
		fmt.Printf("[-] "+time.Now().Format("2006-01-02 15:04:05")+" Simhash of %s is %x\n", single_url, hashes[i])
	}
	fmt.Printf("Simhash比对中 ...\n")
	ResMap := make(map[string]bool)
	for k, _ := range urls {
		ResMap[string(urls[k])] = true
	}
	for i, _ := range urls {
		for j := i + 1; j < len(urls); j++ {
			cp_res := similarity(hashes[i], hashes[j])
			//fmt.Printf("Comparison of `%s` and `%s`: %.5f%% \n", urls[i], urls[j], cp_res)
			if (cp_res/100 - similar) > epsilon {
				ResMap[string(urls[j])] = false
			}
		}
	}
	fmt.Printf("结果保存中 ...\n")
	// 遍历linesMap
	for line, istrue := range ResMap {
		if istrue {
			line = fmt.Sprintf(line + "\n")
			saveFile(savePath, line)
		}
	}
	fmt.Printf("结果保存至" + savePath)
}

func saveFile(savePath, str string) {
	file, err := os.OpenFile(savePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 将字符串写入文件
	_, err = file.WriteString(str)
	if err != nil {
		panic(err)
	}
}

func inputFile(filePath string) ([][]byte, error) {
	// 打开文本文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	var urls [][]byte

	// 逐行读取文件内容
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// 获取每行的 URL
		url := scanner.Text()

		// 将URL转换为所需的格式
		formattedURL := []byte(url)

		// 添加到切片中
		urls = append(urls, formattedURL)
	}

	// 检查扫描过程中是否有错误
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file: %v", err)
	}

	return urls, nil
}

// 百分比计算
func similarity(a uint64, b uint64) float64 {
	percent := Compare(a, b)
	return 100 - (float64(percent)/64.0)*100
}

type Vector [64]float64

// Feature consists of a 64-bit hash and a weight
type Feature interface {
	// Sum returns the 64-bit sum of this feature
	Sum() uint64

	// Weight returns the weight of this feature
	Weight() float64
}

// FeatureSet represents a set of features in a given document
type FeatureSet interface {
	GetFeatures() []Feature
}

// Vectorize generates 64 dimension vectors given a set of features.
// Vectors are initialized to zero. The i-th element of the vector is then
// incremented by weight of the i-th feature if the i-th bit of the feature
// is set, and decremented by the weight of the i-th feature otherwise.
func Vectorize(features []Feature) Vector {
	var v Vector
	//遍历features里每个单词的feature
	for _, feature := range features {
		//获取单个单词的Sum
		sum := feature.Sum()
		//获取单个单词的Weight
		weight := feature.Weight()
		//64次循环
		for i := uint8(0); i < 64; i++ {
			//依次获取由大到小每一位二进制位
			bit := ((sum >> i) & 1)
			//如果该店bit值为1，则该位权重增加
			if bit == 1 {
				v[i] += weight
			} else {
				v[i] -= weight
			}
		}
	}
	return v
}

// VectorizeBytes generates 64 dimension vectors given a set of [][]byte,
// where each []byte is a feature with even weight.
//
// Vectors are initialized to zero. The i-th element of the vector is then
// incremented by weight of the i-th feature if the i-th bit of the feature
// is set, and decremented by the weight of the i-th feature otherwise.
func VectorizeBytes(features [][]byte) Vector {
	var v Vector
	h := fnv.New64()
	for _, feature := range features {
		h.Reset()
		h.Write(feature)
		sum := h.Sum64()
		for i := uint8(0); i < 64; i++ {
			bit := ((sum >> i) & 1)
			if bit == 1 {
				v[i]++
			} else {
				v[i]--
			}
		}
	}
	return v
}

// Fingerprint returns a 64-bit fingerprint of the given vector.
// The fingerprint f of a given 64-dimension vector v is defined as follows:
//
//	f[i] = 1 if v[i] >= 0
//	f[i] = 0 if v[i] < 0
func Fingerprint(v Vector) uint64 {
	var f uint64
	for i := uint8(0); i < 64; i++ {
		if v[i] >= 0 {
			f |= (1 << i)
		}
	}
	return f
}

type feature struct {
	sum    uint64
	weight float64
}

// Sum returns the 64-bit hash of this feature
func (f feature) Sum() uint64 {
	return f.sum
}

// Weight returns the weight of this feature
func (f feature) Weight() float64 {
	return f.weight
}

// Returns a new feature representing the given byte slice, using a weight of 1
func NewFeature(f []byte) feature {
	h := fnv.New64()
	h.Write(f)
	return feature{h.Sum64(), 1}
}

// Returns a new feature representing the given byte slice with the given weight
func NewFeatureWithWeight(f []byte, weight float64) feature {
	fw := NewFeature(f)
	fw.weight = weight
	return fw
}

// Compare calculates the Hamming distance between two 64-bit integers
//
// Currently, this is calculated using the Kernighan method [1]. Other methods
// exist which may be more efficient and are worth exploring at some point
//
// [1] http://graphics.stanford.edu/~seander/bithacks.html#CountBitsSetKernighan
func Compare(a uint64, b uint64) uint8 {
	v := a ^ b
	var c uint8
	for c = 0; v != 0; c++ {
		v &= v - 1
	}
	return c
}

// Returns a 64-bit simhash of the given feature set
func Simhash(fs FeatureSet) uint64 {
	return Fingerprint(Vectorize(fs.GetFeatures()))
}

// Returns a 64-bit simhash of the given bytes
func SimhashBytes(b [][]byte) uint64 {
	return Fingerprint(VectorizeBytes(b))
}

// WordFeatureSet is a feature set in which each word is a feature,
// all equal weight.
type WordFeatureSet struct {
	b []byte
}

func NewWordFeatureSet(b []byte) *WordFeatureSet {
	fs := &WordFeatureSet{b}
	fs.normalize()
	return fs
}

func (w *WordFeatureSet) normalize() {
	w.b = bytes.ToLower(w.b)
}

var boundaries = regexp.MustCompile(`[\w']+(?:\://[\w\./]+){0,1}`)
var unicodeBoundaries = regexp.MustCompile(`[\pL-_']+`)

// Returns a []Feature representing each word in the byte slice
func (w *WordFeatureSet) GetFeatures() []Feature {
	return getFeatures(w.b, boundaries)
}

// UnicodeWordFeatureSet is a feature set in which each word is a feature,
// all equal weight.
//
// See: http://blog.golang.org/normalization
// See: https://groups.google.com/forum/#!topic/golang-nuts/YyH1f_qCZVc
type UnicodeWordFeatureSet struct {
	b []byte
	f norm.Form
}

func NewUnicodeWordFeatureSet(b []byte, f norm.Form) *UnicodeWordFeatureSet {
	fs := &UnicodeWordFeatureSet{b, f}
	fs.normalize()
	return fs
}

func (w *UnicodeWordFeatureSet) normalize() {
	b := bytes.ToLower(w.f.Append(nil, w.b...))
	w.b = b
}

// Returns a []Feature representing each word in the byte slice
func (w *UnicodeWordFeatureSet) GetFeatures() []Feature {
	return getFeatures(w.b, unicodeBoundaries)
}

// Splits the given []byte using the given regexp, then returns a slice
// containing a Feature constructed from each piece matched by the regexp
func getFeatures(b []byte, r *regexp.Regexp) []Feature {
	//将原有[]byte分为单词数组
	words := r.FindAll(b, -1)
	//根据单词数量创建空间
	features := make([]Feature, len(words))
	for i, w := range words {
		//获取每个单词的feature
		//feature默认有两个参数(Value,Weight)
		features[i] = NewFeature(w)
	}
	//一个句子返回一个features
	return features
}

// Shingle returns the w-shingling of the given set of bytes. For example, if the given
// input was {"this", "is", "a", "test"}, this returns {"this is", "is a", "a test"}
func Shingle(w int, b [][]byte) [][]byte {
	if w < 1 {
		// TODO: use error here instead of panic?
		panic("simhash.Shingle(): k must be a positive integer")
	}

	if w == 1 {
		return b
	}

	if w > len(b) {
		w = len(b)
	}

	count := len(b) - w + 1
	shingles := make([][]byte, count)
	for i := 0; i < count; i++ {
		shingles[i] = bytes.Join(b[i:i+w], []byte(" "))
	}
	return shingles
}

type pw struct {
	Value  string
	Weight float64
}

type urlt struct {
	Host     pw
	Path     pw
	RawQuery pw
	Fragment pw
	Scheme   pw
}

// func getFeaturesFromURI(uri string) ([]Feature, error) {
func GetFeaturesFromURI(uri string, weight_arr []string) []Feature {
	parse, err := url.Parse(uri)
	if err != nil {
		return nil
	}
	urlWeights := Setutval_wei(parse.Host, parse.Path, parse.RawQuery, parse.Fragment, parse.Scheme, weight_arr)
	//fmt.Printf("%s %s %s %s %s",parse.Host,parse.Path,parse.RawQuery,parse.Fragment,parse.Scheme)

	//处理url
	urlWeights.Path.Value = strings.ReplaceAll(urlWeights.Path.Value, "//", "/")
	_, urlWeights.Path.Value, _ = strings.Cut(urlWeights.Path.Value, "/")
	urlWeights.RawQuery.Value = strings.ReplaceAll(urlWeights.RawQuery.Value, "&&", "&")
	//路径分割、参数分割
	pathSplit := strings.Split(urlWeights.Path.Value, "/")
	paramSplit := strings.Split(urlWeights.RawQuery.Value, "&")

	//两块小N权重计算
	pathWeight := calculateWeight(urlWeights.Path.Weight, len(pathSplit))
	paramWeight := calculateWeight(urlWeights.RawQuery.Weight, len(paramSplit))

	//返回结果初始化
	//文章方法默认没有加上Fragment
	features := make([]Feature, 0, len(pathSplit)+len(paramSplit)+2)
	appendFeature := func(val string, weight float64) {
		features = append(features, NewFeatureWithWeight([]byte(val), weight))
	}
	//加入元素
	appendFeature(urlWeights.Scheme.Value, urlWeights.Scheme.Weight)
	appendFeature(urlWeights.Host.Value, urlWeights.Host.Weight)

	//2块小N
	for _, value := range pathSplit {
		appendFeature(value, pathWeight)
	}

	for _, value := range paramSplit {
		appendFeature(value, paramWeight)
	}
	return features
}

func Setutval_wei(val1, val2, val3, val4, val5 string, weight_arr []string) urlt {
	var num_arr [5]float64
	for i, s := range weight_arr {
		num, err := strconv.ParseFloat(s, 64)
		if err != nil {
			fmt.Println("请输入正确的权重形式。")
		}
		num_arr[i] = num
	}

	return urlt{
		Host: pw{
			Value:  val1,
			Weight: num_arr[0],
		},
		Path: pw{
			Value:  val2,
			Weight: num_arr[1],
		},
		RawQuery: pw{
			Value:  val3,
			Weight: num_arr[2],
		},
		Fragment: pw{
			Value:  val4,
			Weight: num_arr[3],
		},
		Scheme: pw{
			Value:  val5,
			Weight: num_arr[4],
		},
	}
}

// 算权重
func calculateWeight(totalWeight float64, partsCount int) float64 {
	if partsCount > 0 {
		return totalWeight / float64(partsCount)
	}
	return totalWeight
}
