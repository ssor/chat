package server

import (
	"time"

	"github.com/ssor/chat/node/server/hub"
)

type SummaryReport struct {
	UserCount       int                    `json:"userTotal"`
	UserOnlineCount int                    `json:"userOnline"`
	DataTraficIn    int64                  `json:"dataIn"`
	DataTraficOut   int64                  `json:"dataOut"`
	Reports         []*InvestigationReport `json:"reports"`
}

func MakeHubsStatusReport(interviewees map[string]*hub.Hub) *SummaryReport {
	reports := []*InvestigationReport{}
	recycleCh := make(RecycleChan, len(interviewees))

	// for _, wee := range interviewees {
	// 	go wee.acceptInterview(&Questionnaire{recycleCh})
	// }

	f := func() {
		allReportRecycled := false
		for {
			if allReportRecycled {
				break
			}

			select {
			case report := <-recycleCh:
				reports = append(reports, report)
				delete(interviewees, report.ID)
				if len(interviewees) <= 0 {
					allReportRecycled = true
				}
			case <-time.After(time.Second * 60):
				close(recycleCh)
				allReportRecycled = true
			}
		}
	}
	f()

	sr := &SummaryReport{}
	for _, report := range reports {
		sr.DataTraficIn += report.DataTraficIn
		sr.DataTraficOut += report.DataTraficOut
		sr.UserCount += report.UserCount
		sr.UserOnlineCount += report.UserOnlineCount
	}

	if len(interviewees) > 0 {
		for id := range interviewees {
			report := NewInvestigationReport(id, -1, -1, -1, -1)
			reports = append(reports, report)
		}

	}
	sr.Reports = reports
	return sr
}

type InvestigationReport struct {
	ID              string `json:"id"`
	UserCount       int    `json:"userTotal"`
	UserOnlineCount int    `json:"userOnline"`
	DataTraficIn    int64  `json:"dataIn"`
	DataTraficOut   int64  `json:"dataOut"`
}

func NewInvestigationReport(id string, userCount, online int, in, out int64) *InvestigationReport {
	return &InvestigationReport{
		ID:              id,
		UserCount:       userCount,
		UserOnlineCount: online,
		DataTraficIn:    in,
		DataTraficOut:   out,
	}
}

type RecycleChan chan *InvestigationReport

type Questionnaire struct {
	RecycleChan
}
