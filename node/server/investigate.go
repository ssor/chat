package server

import "time"

type SummaryReport struct {
	UserCount       int                    `json:"userTotal"`
	UserOnlineCount int                    `json:"userOnline"`
	DataTraficIn    int64                  `json:"dataIn"`
	DataTraficOut   int64                  `json:"dataOut"`
	Reports         []*InvestigationReport `json:"reports"`
}

func MakeHubsStatusReport(interviewees map[string]*Hub) *SummaryReport {
	reports := []*InvestigationReport{}
	recycleCh := make(RecycleChan, len(interviewees))

	for _, wee := range interviewees {
		go wee.acceptInterview(&Questionnaire{recycleCh})
	}

	f := func() {
		all_report_recycled := false
		for {
			if all_report_recycled {
				break
			}

			select {
			case report := <-recycleCh:
				reports = append(reports, report)
				delete(interviewees, report.ID)
				if len(interviewees) <= 0 {
					all_report_recycled = true
				}
			case <-time.After(time.Second * 60):
				close(recycleCh)
				all_report_recycled = true
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

func NewInvestigationReport(id string, user_count, online int, in, out int64) *InvestigationReport {
	return &InvestigationReport{
		ID:              id,
		UserCount:       user_count,
		UserOnlineCount: online,
		DataTraficIn:    in,
		DataTraficOut:   out,
	}
}

type RecycleChan chan *InvestigationReport

type Questionnaire struct {
	RecycleChan
}

// //调查员
// type Investigator struct {
// 	lastInvestComplete bool
// }

// func NewInvestigator() *Investigator {
// 	return &Investigator{
// 		lastInvestComplete: true,
// 	}
// }

// //发起调查
// //发起调查时会针对各个对象发放调查问卷,并在指定时间内回收
// //结束回收问卷有两种情况:1,所有问卷均以回收完毕;2,超过最长回收时间
// func (in *Investigator) StartInvestigation(interviewees map[string]*Hub) {
// 	// if in.lastInvestComplete == false {
// 	// 	return
// 	// }
// 	// in.lastInvestComplete = false
// 	reports := []*InvestigationReport{}
// 	recycleCh := make(RecycleChan, len(interviewees))

// 	for _, wee := range interviewees {
// 		go wee.acceptInterview(&Questionnaire{recycleCh})
// 	}

// 	go func() {
// 		all_report_recycled := false
// 		for {
// 			if all_report_recycled {
// 				// in.lastInvestComplete = true
// 				// in.ReportResult(reports, interviewees)
// 				break
// 			}

// 			select {
// 			case report := <-recycleCh:
// 				reports = append(reports, report)
// 				delete(interviewees, report.ID)
// 				// deleteInterviewee(interviewees, report.ID)
// 				if len(interviewees) <= 0 {
// 					all_report_recycled = true
// 				}
// 			case <-time.After(time.Second * 60):
// 				close(recycleCh)
// 				all_report_recycled = true
// 			}
// 		}
// 	}()
// }

// func (in *Investigator) ReportResult(reports []*InvestigationReport, intervieweesLeft map[string]*Hub) {
// 	if len(intervieweesLeft) <= 0 {
// 		log.InfoF("%d groups running well", len(reports))
// 	} else {
// 		log.SysF("%d groups no running state report ", len(intervieweesLeft))
// 	}

// 	report_total := InvestigationReport{}
// 	for _, report := range reports {
// 		report_total.DataTraficIn += report.DataTraficIn
// 		report_total.DataTraficOut += report.DataTraficOut
// 		report_total.UserCount += report.UserCount
// 		report_total.UserOnlineCount += report.UserOnlineCount
// 	}

// 	log.InfoF("Users all: %d | Online: %d | Bytes In: %d | Out: %d", report_total.UserCount, report_total.UserOnlineCount, report_total.DataTraficIn, report_total.DataTraficOut)
// }
