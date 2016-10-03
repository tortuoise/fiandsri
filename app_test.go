package fiandsri

import (

        "fmt"
        "net/http"
        "image"
        "image/color"
        "image/draw"
        "image/jpeg"
        _"image/png"
        "os"
        "strconv"
        "time"
	_ "google.golang.org/api/calendar/v3"
        "github.com/tortuoise/wattwerks/utils"
)

func main() {

        //grid := 35
        rows := 5
        cols := 7
        mult := 50
        now := time.Now() // today's date
        ven, err := time.LoadLocation("IST")
        if err != nil {
                ven, _= time.LoadLocation("UTC")
        }
        dday := time.Date(2016, time.November, 02, 8, 30, 0, 0, ven)

        ds := 0
        for ds = 0; now.Before(dday); ds++{
                now = now.Add(time.Hour * 24)
        }
        fmt.Println(ds, " days to go")

        this_mth := time.Now().Month()
        //next_mth := this_mth + 1
        //nxxt_mth :=  this_mth + 2
        fmt.Println(utils.DayspMonth(this_mth))

        first := time.Date(now.Year(), this_mth, 1, 12, 0, 0, 0, ven).Weekday()
        ifirst := int(first)

        mth, err := CreateImage(this_mth, ifirst, rows, cols, "one.jpg", mult)
        if err != nil {
                fmt.Println(err)
        }
        if err := WriteImage(mth, "g127.jpg"); err != nil {
                fmt.Println(err)
        }

        ci := NewCalImage(time.Now(), dday, rows, cols)
        err = ci.CreateImages()
        if err != nil {
                fmt.Println(err)
        }
        for i, img := range ci.images {
                err = WriteImage(img, "f"+strconv.Itoa(i)+".jpg")
                if err != nil {
                        fmt.Println(err)
                }
        }
}


//CreateImage returns generated image for this_mth (starting on ifirst time.Weekday) in grid of rows * cols with each subimg sourced from file with known size
func CreateImage(this_mth time.Month, ifirst, rows, cols int, subimg string, subimgsz int) (*image.Gray, error) {

        rdr, err := os.Open("one.jpg")
        if err != nil {
                return nil, err
        }
        defer rdr.Close()
        digit, _, err := image.Decode(rdr)
        if err != nil {
                return nil, err
        }
        digitg := utils.Convert(digit, color.GrayModel.Convert)
        bl := digitg.Bounds()


        mult := subimgsz
        grid := rows * cols
        x, y, p := 0, 0, 0
        days := make([]*image.Gray, 0, grid)
        mth := image.NewGray(image.Rect(0,0, cols * mult, rows * mult))
        for n:=1; n <= grid; n++ {
                digitgg := utils.Copy(digitg)
                if t := (n % cols); t != 0 {
                        x = (n % cols) * mult
                } else {
                        x = cols * mult
                }
                if t := (n / cols); n % cols != 0 {
                        y = (t + 1)  * mult
                } else {
                        y = t  * mult
                }
                day := mth.SubImage(image.Rect(x-mult, y-mult, x, y)).(*image.Gray)
                days = append(days, day)
                grey := color.Gray{7 *uint8( n)}
                for yy := bl.Min.Y; yy < bl.Max.Y; yy++ {
                        for xx := bl.Min.X; xx < bl.Max.X; xx++ {
                                if gy := digitgg.At(xx,yy).(color.Gray).Y; uint16(gy) > 100 { //!= color.White.Y {
                                        //digitg.SetGray(xx, yy, grey)
                                        digitgg.SetGray(xx, yy, grey)
                                }
                        }
                }
                if n > ifirst && p < utils.DayspMonth(this_mth){
                        draw.Draw(day, day.Bounds(), digitgg, image.ZP, draw.Src)
                        p++
                } else {
                        draw.Draw(day, day.Bounds(), &image.Uniform{color.White}, image.ZP, draw.Src)
                }
        }
        return mth, nil
        //draw.Draw(mth, mth.Bounds(), &image.Uniform{grey}, image.ZP, draw.Src)

}

//WriteImage writes the image provided as argument to the file with name provided as argument
func WriteImage(mth *image.Gray, filename string) error {

        f, err := os.Create(filename) //("g127.jpg")
        if err != nil {
                return err
        }
        defer f.Close()
        err = jpeg.Encode(f, mth, nil)
        if err != nil {
                return err
        }
        return nil

}


func (ci *CalImage) CreateImages() error {

        rdr, err := os.Open("1.jpg")
        if err != nil {
                return err
        }
        digit, _, err := image.Decode(rdr)
        if err != nil {
                return err
        }
        rdr.Close()
        digitg := utils.Convert(digit, color.GrayModel.Convert)
        bl := digitg.Bounds()

        mult := bl.Max.Y
        for ii, n35 := range ci.n35 {
                rows := ci.rows
                cols := ci.cols
                grid := rows * cols
                if n35[0] < 0 {
                        rows++
                        grid = rows * cols
                        n35[1] = n35[1] - n35[0]
                        n35[0] += 7
                        fmt.Println(n35)
                }
                x, y, p := 0, 0, 1
                days := make([]*image.Gray, 0, grid)
                mth := image.NewGray(image.Rect(0,0, cols * mult, rows * mult))
                for n:=1; n <= grid; n++ { //n 1...35/42

                        //digitgg := utils.Copy(digitg)
                        if t := (n % cols); t != 0 {
                                x = (n % cols) * mult
                        } else {
                                x = cols * mult
                        }
                        if t := (n / cols); n % cols != 0 {
                                y = (t + 1)  * mult
                        } else {
                                y = t  * mult
                        }
                        day := mth.SubImage(image.Rect(x-mult, y-mult, x, y)).(*image.Gray)
                        days = append(days, day)
                        if n35[0] > 0 && n > n35[0] && p <= utils.DayspMonth(ci.start.Month() + time.Month(ii)){
                                rdr, err := os.Open(strconv.Itoa(p) + ".jpg")
                                if err != nil {
                                        return err
                                }
                                digit, _, err := image.Decode(rdr)
                                if err != nil {
                                        return err
                                }
                                rdr.Close()
                                digitg := utils.Convert(digit, color.GrayModel.Convert)
                                digitgg := utils.Copy(digitg)
                                grey := color.Gray{5 *uint8( n)}
                                for yy := bl.Min.Y; yy < bl.Max.Y; yy++ {
                                        for xx := bl.Min.X; xx < bl.Max.X; xx++ {
                                                if gy := digitgg.At(xx,yy).(color.Gray).Y; uint16(gy) > 100 { //!= color.White.Y {
                                                        //digitg.SetGray(xx, yy, grey)
                                                        digitgg.SetGray(xx, yy, grey)
                                                }
                                        }
                                }
                                draw.Draw(day, day.Bounds(), digitgg, image.ZP, draw.Src)
                                p++
                        } else {
                                draw.Draw(day, day.Bounds(), &image.Uniform{color.White}, image.ZP, draw.Src)
                        }
                }
                //return mth, nil
                //ci.images = append(ci.images, mth)
                ci.images[ii] = mth
        }
        return nil
        //draw.Draw(mth, mth.Bounds(), &image.Uniform{grey}, image.ZP, draw.Src)

}

type CalImage struct {

        ven *time.Location
        start time.Time
        end time.Time
        now time.Time
        images []*image.Gray
        cols int
        rows int
        n35 [][]int // stores the start and end (int 0<n<=35) on grid of each month. If end < start, start is on previous month's grid to avoid overflow

}

func NewCalImage(st, nd time.Time, rs, cs int) *CalImage {

        x := newCalImage(st, nd, rs, cs)
        for i, ov := range x.Overflow() {
                xst := x.start.Month()
                xsti := xst + time.Month(i)
                if ov != 0 {
                        x.n35[i] = append(x.n35[i], ov )
                        x.n35[i] = append(x.n35[i], x.n35[i][0] + utils.DayspMonth(xsti))// - 1)
                } else {
                        xs := make([]int, 2)
                        xs[0] = int(time.Date(x.start.Year(), xsti, 1, 12, 0, 0, 0, x.ven).Weekday()) //+ 1
                        xs[1] = xs[0] + utils.DayspMonth(xsti) //- 1
                        x.n35[i] = xs//append(x.n35[i], int(time.Date(x.start.Year(), x.start.Month(), 1, 12, 0, 0, 0, x.ven).Weekday()))
                        //x.n35[i] = append(x.n35[i], x.n35[i][0] + utils.DayspMonth(x.start.Month()))
                }
        }
        return x

}

func newCalImage(st, nd time.Time, rs,cs int) *CalImage {

        venue, _ := time.LoadLocation("UTC")
        if st.Year() == nd.Year() {
                per := nd.Month() - st.Month() + 1
                return &CalImage{ven: venue, start:st, end:nd, cols:cs, rows:rs, n35: make([][]int, per, per), images: make([]*image.Gray, per, per)}
        } else {
                return &CalImage{ven: venue, start:st, end:nd, cols:cs, rows:rs}
        }

}

//Overflow checks whether any of the months in the period overflows a 7x5 grid. For instance if 31 day month starts on Friday 
func (ci *CalImage) Overflow() []int {

        mths := make([]int,0)
        if ci.OneYear() {
                for n := ci.start.Month(); n <= ci.end.Month(); n++ {
                        nd := utils.DayspMonth(n)
                        first := int(time.Date(ci.start.Year(), n, 1, 12, 0, 0, 0, ci.ven).Weekday()) + 1 // Weekday starts at Sunday = 0
                        if nd + first > 35 {
                                mths = append(mths, first - 8)
                        } else {
                                mths = append(mths, 0)
                        }
                }
                return mths
        } else {
                return mths
        }
}

//OneYear checks whether start and end are in same year 
func (ci *CalImage) OneYear() bool {

        return ci.start.Year() == ci.end.Year()

}

