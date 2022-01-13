package metadata

import (
	"path/filepath"

	"github.com/checkmarxDev/ast-sast-export/internal/app/interfaces"
	"github.com/checkmarxDev/ast-sast-export/internal/app/report"
	"github.com/checkmarxDev/ast-sast-export/internal/app/worker"
	"github.com/checkmarxDev/ast-sast-export/internal/integration/similarity"
	"github.com/pkg/errors"
)

type MetadataProvider interface {
	GetMetadataRecord(scanID string, queries []*Query) (*Record, error)
}

type MetadataFactory struct {
	astQueryIDProvider   interfaces.ASTQueryIDRepo
	similarityIDProvider similarity.SimilarityIDProvider
	sourceProvider       interfaces.SourceFileRepo
	methodLineProvider   interfaces.MethodLineRepo
	tmpDir               string
}

func NewMetadataFactory(
	astQueryIDProvider interfaces.ASTQueryIDRepo,
	similarityIDProvider similarity.SimilarityIDProvider,
	sourceProvider interfaces.SourceFileRepo,
	methodLineProvider interfaces.MethodLineRepo,
	tmpDir string,
) *MetadataFactory {
	return &MetadataFactory{
		astQueryIDProvider,
		similarityIDProvider,
		sourceProvider,
		methodLineProvider,
		tmpDir,
	}
}

func (e *MetadataFactory) GetMetadataRecord(scanID string, queries []*Query) (*Record, error) {
	output := &Record{Queries: []*RecordQuery{}}

	for queryIdx, query := range queries {
		output.Queries = append(output.Queries, &RecordQuery{QueryID: query.QueryID})
		astQueryID, astQueryIDErr := e.astQueryIDProvider.GetQueryID(query.Language, query.Name, query.Group)
		if astQueryIDErr != nil {
			// maybe the group changed
			queryList, queryListErr := e.astQueryIDProvider.GetAllQueryIDsByGroup(query.Language, query.Name)
			if queryListErr != nil {
				return nil, errors.Wrap(astQueryIDErr, "could not get AST query ids by group")
			}
			if len(queryList) == 0 {
				return nil, errors.Wrap(astQueryIDErr, "could not get AST query id")
			}
			if len(queryList) > 1 {
				return nil, errors.Wrapf(
					astQueryIDErr,
					"could not get AST query id - found more than one query for language %s and name %s",
					query.Language,
					query.Name,
				)
			}
			astQueryID = queryList[0].QueryID
		}
		methodLinesByPath, methodLineErr := e.methodLineProvider.GetMethodLinesByPath(scanID, query.QueryID)
		if methodLineErr != nil {
			return nil, errors.Wrap(methodLineErr, "could not get method lines")
		}
		var filesToDownload []interfaces.SourceFile
		for _, result := range query.Results {
			if ok1 := findSourceFile(result.FirstNode.FileName, filesToDownload); ok1 == nil {
				filesToDownload = append(filesToDownload, interfaces.SourceFile{
					RemoteName: result.FirstNode.FileName,
					LocalName:  filepath.Join(e.tmpDir, result.FirstNode.FileName),
				})
			}
			if ok2 := findSourceFile(result.LastNode.FileName, filesToDownload); ok2 == nil {
				filesToDownload = append(filesToDownload, interfaces.SourceFile{
					RemoteName: result.LastNode.FileName,
					LocalName:  filepath.Join(e.tmpDir, result.LastNode.FileName),
				})
			}
		}
		downloadErr := e.sourceProvider.DownloadSourceFiles(scanID, filesToDownload)
		if downloadErr != nil {
			return nil, errors.Wrap(downloadErr, "could not download source code")
		}

		// produce calculation jobs
		similarityCalculationJobs := make(chan SimilarityCalculationJob)
		q := query
		go func() {
			for _, result := range q.Results {
				firstSourceFile := findSourceFile(result.FirstNode.FileName, filesToDownload)
				lastSourceFile := findSourceFile(result.LastNode.FileName, filesToDownload)
				methodLines := findResultPath(result.PathID, methodLinesByPath).MethodLines
				similarityCalculationJobs <- SimilarityCalculationJob{
					result.ResultID, result.PathID,
					firstSourceFile.LocalName, result.FirstNode.Name, result.FirstNode.Line, result.FirstNode.Column, methodLines[0],
					lastSourceFile.LocalName, result.LastNode.Name, result.LastNode.Line, result.LastNode.Column, methodLines[len(methodLines)-1],
					astQueryID,
				}
			}
			close(similarityCalculationJobs)
		}()

		// consume calculation jobs
		similarityCalculationResults := make(chan SimilarityCalculationResult, len(query.Results))
		for consumerID := 1; consumerID <= worker.GetNumCPU(); consumerID++ {
			go func() {
				for job := range similarityCalculationJobs {
					similarityID, similarityIDErr := e.similarityIDProvider.Calculate(
						job.Filename1, job.Name1, job.Line1, job.Column1, job.MethodLine1,
						job.Filename2, job.Name2, job.Line2, job.Column2, job.MethodLine2,
						job.QueryID,
					)
					similarityCalculationResults <- SimilarityCalculationResult{
						ResultID:     job.ResultID,
						PathID:       job.PathID,
						SimilarityID: similarityID,
						Err:          similarityIDErr,
					}
				}
			}()
		}

		// handle calculation results
		for range query.Results {
			r := <-similarityCalculationResults
			if r.Err != nil {
				return nil, r.Err
			}
			var recordResult *RecordResult
			for _, x := range output.Queries[queryIdx].Results {
				if x.ResultID == r.ResultID {
					recordResult = x
					break
				}
			}
			if recordResult == nil {
				recordResult = &RecordResult{ResultID: r.ResultID}
				output.Queries[queryIdx].Results = append(output.Queries[queryIdx].Results, recordResult)
			}
			var recordPath *RecordPath
			for _, x := range recordResult.Paths {
				if x.PathID == r.PathID {
					recordPath = x
					break
				}
			}
			if recordPath == nil {
				recordPath = &RecordPath{PathID: r.PathID, SimilarityID: r.SimilarityID}
				recordResult.Paths = append(recordResult.Paths, recordPath)
			}
		}
	}

	return output, nil
}

func findSourceFile(remoteName string, sourceFiles []interfaces.SourceFile) *interfaces.SourceFile {
	for _, v := range sourceFiles {
		if v.RemoteName == remoteName {
			return &v
		}
	}
	return nil
}

func findResultPath(pathID string, methodLines []*interfaces.ResultPath) *interfaces.ResultPath {
	for _, v := range methodLines {
		if v.PathID == pathID {
			return v
		}
	}
	return nil
}

func GetQueriesFromReport(reportReader *report.CxXMLResults) []*Query {
	var output []*Query
	for i := 0; i < len(reportReader.Queries); i++ {
		q := reportReader.Queries[i]
		query := &Query{
			QueryID:  q.ID,
			Name:     q.Name,
			Language: q.Language,
			Group:    q.Group,
		}
		for j := 0; j < len(q.Results); j++ {
			r := q.Results[j]
			// only triaged results will have metadata records generated
			if r.Remark == "" {
				continue
			}
			for k := 0; k < len(r.Paths); k++ {
				p := r.Paths[k]
				firstNode := p.PathNodes[0]
				lastNode := p.PathNodes[len(p.PathNodes)-1]
				query.Results = append(query.Results, &Result{
					ResultID: p.ResultID,
					PathID:   p.PathID,
					FirstNode: Node{
						FileName: firstNode.FileName,
						Name:     firstNode.Name,
						Line:     firstNode.Line,
						Column:   firstNode.Column,
					},
					LastNode: Node{
						FileName: lastNode.FileName,
						Name:     lastNode.Name,
						Line:     lastNode.Line,
						Column:   lastNode.Column,
					},
				})
			}
		}
		if len(query.Results) > 0 {
			output = append(output, query)
		}
	}
	return output
}