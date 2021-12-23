package metadata

import (
	"path/filepath"

	"github.com/checkmarxDev/ast-sast-export/internal/app/interfaces"

	"github.com/checkmarxDev/ast-sast-export/internal/app/worker"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/similarity"
	"github.com/pkg/errors"
)

type MetadataProvider interface {
	GetMetadataRecords(scanID string, query *Query) ([]*Record, error)
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

func (e *MetadataFactory) GetMetadataRecords(scanID string, query *Query) ([]*Record, error) {
	astQueryID, astQueryIDErr := e.astQueryIDProvider.GetQueryID(query.Language, query.Name, query.Group)
	if astQueryIDErr != nil {
		return nil, errors.Wrap(astQueryIDErr, "could not get AST query id")
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
	var output []*Record

	// produce calculation jobs
	similarityCalculationJobs := make(chan SimilarityCalculationJob)
	go func() {
		for _, result := range query.Results {
			firstSourceFile := findSourceFile(result.FirstNode.FileName, filesToDownload)
			lastSourceFile := findSourceFile(result.LastNode.FileName, filesToDownload)
			methodLines := methodLinesByPath[result.PathID]
			similarityCalculationJobs <- SimilarityCalculationJob{
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
					SimilarityID: similarityID,
					Err:          similarityIDErr,
				}
			}
		}()
	}

	// handle calculation results
	for _, result := range query.Results {
		r := <-similarityCalculationResults
		if r.Err != nil {
			return nil, r.Err
		}
		output = append(output, &Record{
			QueryID:      query.QueryID,
			SimilarityID: r.SimilarityID,
			PathID:       result.PathID,
			ResultID:     result.ResultID,
		})
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
