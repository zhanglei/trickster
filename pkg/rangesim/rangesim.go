/**
* Copyright 2018 Comcast Cable Communications Management, LLC
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
* http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

// Package rangesim is a sample HTTP server that fully supports HTTP Range Requests
// it is used by Trickster for unit testing and integration testing
package rangesim

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

// Response Object Data Constants
const contentLength = int64(1222)

// Jan 1 2020 00:00:00 GMT
var lastModified = time.Unix(1577836800, 0)

const contentType = "text/plain; charset=utf-8"
const separator = "TestRangeServerBoundary"
const maxAge = "max-age=60"

// Body is the body that RangeSim uses to serve content
const Body = `Lorem ipsum dolor sit amet, mel alia definiebas ei, labore eligendi ` +
	`signiferumque id sed. Dico tantas fabulas et vel, maiorum splendide has an. Te mea ` +
	`suas commune concludaturque. Qui feugait tacimates te.

` + `Ea sea error altera efficiantur, ex possit appetere eum. Sed cu sanctus blandit definiebas, ` +
	`movet accumsan no mei. Vim diam molestie singulis cu, et sanctus appetere ius, his ut ` +
	`consulatu vituperata. Graece graeco sit ut, an quem summo splendide duo. Iisque ` +
	`sapientem interpretaris pro ad, alii mazim pro te. Malis laoreet facilis sea te. An ` +
	`ferri albucius vel, altera volumus legendos has in.

` + `His ne dolore rationibus. Ut qui ferri malorum. Mel commune atomorum cu. Ut mollis ` +
	`reprimique nam, eos quot mutat molestie id. Mea error legere contentiones et, ponderum ` +
	`accusamus est eu. Detraxit repudiandae signiferumque ne eos.

` + `Ius ne periculis consequat, ea usu brute mediocritatem, an qui reque falli deseruisse. ` +
	`Vix ne aeque movet. Novum homero referrentur in est. No mei adhuc malorum.

` + `Pri vitae sapientem ad, qui libris prompta ei. Ne quem fabulas dissentiet cum, error ` +
	`legimus vis cu. Te eum lorem liber aliquando, eirmod diceret vis ad. Eos et facer tation. ` +
	`Etiam phaedrum ea est, an nec summo mediocritatem.`

// HTTP Elements
const hnAcceptRanges = `Accept-Ranges`
const hnCacheControl = `Cache-Control`
const hnContentRange = `Content-Range`
const hnContentType = `Content-Type`
const hnIfModifiedSince = `If-Modified-Since`
const hnLastModified = `Last-Modified`
const hnRange = `Range`

const hvMultipartByteRange = `multipart/byteranges; boundary=`
const byteRequestRangePrefix = "bytes="
const byteResponsRangePrefix = "bytes "

type byteRanges []byteRange

func (brs byteRanges) validate() bool {
	for _, r := range brs {
		if r.start < 0 || r.end >= contentLength || r.end < r.start {
			return false
		}
	}
	return true
}

type byteRange struct {
	start int64
	end   int64
}

func (br byteRange) contentRangeHeader() string {
	return fmt.Sprintf("%s%d-%d/%d", byteResponsRangePrefix, br.start, br.end, contentLength)
}

func (brs byteRanges) writeMultipartResponse(w io.Writer) error {

	mw := multipart.NewWriter(w)
	mw.SetBoundary(separator)
	for _, r := range brs {
		pw, err := mw.CreatePart(
			textproto.MIMEHeader{
				hnContentType:  []string{contentType},
				hnContentRange: []string{r.contentRangeHeader()},
			},
		)
		if err != nil {
			return err
		}
		pw.Write([]byte(Body[r.start : r.end+1]))
	}
	mw.Close()
	return nil
}

func parseRangeHeader(input string) byteRanges {

	if input == "" || !strings.HasPrefix(input, byteRequestRangePrefix) ||
		input == byteRequestRangePrefix {
		return nil
	}
	input = strings.Replace(input, " ", "", -1)[6:]
	parts := strings.Split(input, ",")
	ranges := make(byteRanges, len(parts))

	for i, p := range parts {

		j := strings.Index(p, "-")
		if j < 0 {
			return nil
		}

		var start = int64(-1)
		var end = int64(-1)
		var err error

		if j > 0 {
			start, err = strconv.ParseInt(p[0:j], 10, 64)
			if err != nil {
				return nil
			}
		}

		if j < len(p)-1 {
			end, err = strconv.ParseInt(p[j+1:], 10, 64)
			if err != nil {
				return nil
			}
		}

		ranges[i].start = start
		ranges[i].end = end
	}

	return ranges
}
