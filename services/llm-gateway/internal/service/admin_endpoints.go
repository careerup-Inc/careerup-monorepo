package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tmc/langchaingo/schema"

	pbllm "github.com/careerup-Inc/careerup-monorepo/proto/llm/v1"
)

// IngestDocument ingests a document into the specified collection
func (s *LLMServiceImpl) IngestDocument(ctx context.Context, req *pbllm.IngestDocumentRequest) (*pbllm.IngestDocumentResponse, error) {
	log.Printf("Ingesting document into collection: %s", req.GetCollection())

	// Validate request
	if req.GetContent() == "" {
		return &pbllm.IngestDocumentResponse{
			Success: false,
			Message: "Document content cannot be empty",
		}, nil
	}

	collection := req.GetCollection()
	if collection == "" {
		collection = "default"
	}

	// Generate document ID if not provided
	documentId := req.GetDocumentId()
	if documentId == "" {
		documentId = fmt.Sprintf("doc_%d", time.Now().Unix())
	}

	// Ensure vector store exists for the collection
	if err := s.InitializeVectorStore(collection); err != nil {
		return &pbllm.IngestDocumentResponse{
			DocumentId: documentId,
			Success:    false,
			Message:    fmt.Sprintf("Failed to initialize vector store: %v", err),
		}, nil
	}

	// Create document with metadata
	doc := schema.Document{
		PageContent: req.GetContent(),
		Metadata: map[string]interface{}{
			"document_id": documentId,
			"indexed_at":  time.Now().Format(time.RFC3339),
		},
	}

	// Add custom metadata from request
	for k, v := range req.GetMetadata() {
		doc.Metadata[k] = v
	}

	// Split document into chunks (reusing existing logic)
	chunks := strings.Split(doc.PageContent, "\n\n")
	splitDocs := make([]schema.Document, 0, len(chunks))

	for i, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			continue // Skip empty chunks
		}

		chunkDoc := schema.Document{
			PageContent: strings.TrimSpace(chunk),
			Metadata: map[string]interface{}{
				"document_id": documentId,
				"chunk_index": i,
				"indexed_at":  time.Now().Format(time.RFC3339),
			},
		}

		// Add custom metadata to each chunk
		for k, v := range req.GetMetadata() {
			chunkDoc.Metadata[k] = v
		}

		splitDocs = append(splitDocs, chunkDoc)
	}

	// Add to vector store
	vs := s.vectorStores[collection]
	_, err := vs.AddDocuments(ctx, splitDocs)
	if err != nil {
		return &pbllm.IngestDocumentResponse{
			DocumentId: documentId,
			Success:    false,
			Message:    fmt.Sprintf("Failed to add document to vector store: %v", err),
		}, nil
	}

	log.Printf("Successfully ingested document %s into collection %s with %d chunks",
		documentId, collection, len(splitDocs))

	return &pbllm.IngestDocumentResponse{
		DocumentId:    documentId,
		Success:       true,
		Message:       "Document successfully ingested",
		ChunksCreated: int32(len(splitDocs)),
	}, nil
}

// CreateCollection creates a new vector store collection
func (s *LLMServiceImpl) CreateCollection(ctx context.Context, req *pbllm.CreateCollectionRequest) (*pbllm.CreateCollectionResponse, error) {
	collectionName := req.GetCollectionName()
	log.Printf("Creating collection: %s", collectionName)

	if collectionName == "" {
		return &pbllm.CreateCollectionResponse{
			Success: false,
			Message: "Collection name cannot be empty",
		}, nil
	}

	// Check if collection already exists
	if _, exists := s.vectorStores[collectionName]; exists {
		return &pbllm.CreateCollectionResponse{
			Success:        false,
			Message:        "Collection already exists",
			CollectionName: collectionName,
		}, nil
	}

	// Initialize the vector store for the new collection
	if err := s.InitializeVectorStore(collectionName); err != nil {
		return &pbllm.CreateCollectionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create collection: %v", err),
		}, nil
	}

	log.Printf("Successfully created collection: %s", collectionName)

	return &pbllm.CreateCollectionResponse{
		Success:        true,
		Message:        "Collection created successfully",
		CollectionName: collectionName,
	}, nil
}

// ListCollections lists all available collections
func (s *LLMServiceImpl) ListCollections(ctx context.Context, req *pbllm.ListCollectionsRequest) (*pbllm.ListCollectionsResponse, error) {
	log.Printf("Listing collections")

	var collections []*pbllm.CollectionInfo
	for name := range s.vectorStores {
		collections = append(collections, &pbllm.CollectionInfo{
			Name: name,
			// Note: Chroma doesn't easily provide document count, so we'll leave it empty for now
			// In a production system, you might want to track this separately
		})
	}

	return &pbllm.ListCollectionsResponse{
		Collections: collections,
	}, nil
}

// DeleteCollection deletes a collection and all its documents
func (s *LLMServiceImpl) DeleteCollection(ctx context.Context, req *pbllm.DeleteCollectionRequest) (*pbllm.DeleteCollectionResponse, error) {
	collectionName := req.GetCollectionName()
	log.Printf("Deleting collection: %s", collectionName)

	if collectionName == "" {
		return &pbllm.DeleteCollectionResponse{
			Success: false,
			Message: "Collection name cannot be empty",
		}, nil
	}

	// Check if collection exists
	if _, exists := s.vectorStores[collectionName]; !exists {
		return &pbllm.DeleteCollectionResponse{
			Success: false,
			Message: "Collection does not exist",
		}, nil
	}

	// Remove from our local map
	delete(s.vectorStores, collectionName)

	// TODO Note: Chroma client doesn't provide an easy way to delete collections through the Go client
	// In a production system, you might want to make direct HTTP calls to Chroma's REST API
	// For now, we'll just remove it from our local tracking

	log.Printf("Successfully deleted collection: %s", collectionName)

	return &pbllm.DeleteCollectionResponse{
		Success: true,
		Message: "Collection deleted successfully",
	}, nil
}

// IngestILOData ingests ILO-related data for RAG-enabled career guidance
func (s *LLMServiceImpl) IngestILOData(ctx context.Context, req *pbllm.IngestILODataRequest) (*pbllm.IngestILODataResponse, error) {
	collection := "ilo_career_guidance"
	log.Printf("Ingesting ILO data into collection: %s", collection)

	// Initialize ILO collection if not exists
	if err := s.InitializeVectorStore(collection); err != nil {
		return &pbllm.IngestILODataResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to initialize ILO vector store: %v", err),
		}, nil
	}

	totalDocs := 0

	// Ingest domain descriptions
	domainDocs, err := s.ingestDomainData(ctx, collection)
	if err != nil {
		return &pbllm.IngestILODataResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to ingest domain data: %v", err),
		}, nil
	}
	totalDocs += domainDocs

	// Ingest career mappings and descriptions
	careerDocs, err := s.ingestCareerMappings(ctx, collection)
	if err != nil {
		return &pbllm.IngestILODataResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to ingest career mappings: %v", err),
		}, nil
	}
	totalDocs += careerDocs

	// Ingest career guidance templates
	templateDocs, err := s.ingestGuidanceTemplates(ctx, collection)
	if err != nil {
		return &pbllm.IngestILODataResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to ingest guidance templates: %v", err),
		}, nil
	}
	totalDocs += templateDocs

	log.Printf("Successfully ingested %d ILO documents into collection %s", totalDocs, collection)

	return &pbllm.IngestILODataResponse{
		Success:      true,
		Message:      "ILO data successfully ingested for RAG",
		DocsIngested: int32(totalDocs),
	}, nil
}

func (s *LLMServiceImpl) ingestDomainData(ctx context.Context, collection string) (int, error) {
	// ILO domain knowledge based on the Vietnamese career guidance system
	domains := []struct {
		code, name, description, characteristics string
	}{
		{
			"LANG", "Ngôn ngữ (Language)",
			"Khả năng sử dụng ngôn ngữ, viết lách, giao tiếp và diễn đạt ý tưởng một cách rõ ràng",
			"Người có khả năng ngôn ngữ cao thường thích đọc, viết, nói chuyện, học ngoại ngữ. Họ có thể trở thành nhà văn, nhà báo, biên dịch, giáo viên ngữ văn, luật sư, nhà ngoại giao.",
		},
		{
			"LOGIC", "Logic (Logical)",
			"Tư duy logic, toán học, phân tích dữ liệu và giải quyết vấn đề một cách có hệ thống",
			"Người có tư duy logic mạnh thích làm việc với số liệu, phân tích, tìm ra quy luật. Họ phù hợp với nghề kỹ sư, lập trình viên, nhà toán học, nhà khoa học, kế toán, phân tích tài chính.",
		},
		{
			"DESIGN", "Thiết kế (Design)",
			"Khả năng sáng tạo, thẩm mỹ, thiết kế và tạo ra những sản phẩm có giá trị nghệ thuật",
			"Người có khiếu thiết kế thích tạo ra những thứ đẹp mắt, độc đáo. Họ có thể làm designer, kiến trúc sư, họa sĩ, nhà điêu khắc, nhà thiết kế thời trang, decorator.",
		},
		{
			"MUSIC", "Âm nhạc (Music)",
			"Khả năng cảm nhận, sáng tác và biểu diễn âm nhạc",
			"Người có khiếu âm nhạc có tai nhạy bén, nhịp điệu tốt. Họ có thể trở thành nhạc sĩ, ca sĩ, nhạc công, giáo viên âm nhạc, sound engineer, producer.",
		},
		{
			"MECHANIC", "Cơ khí (Mechanical)",
			"Khả năng hiểu và làm việc với máy móc, thiết bị cơ khí",
			"Người có khiếu cơ khí thích tìm hiểu cách hoạt động của máy móc, sửa chữa, lắp ráp. Họ phù hợp làm kỹ sư cơ khí, thợ máy, kỹ thuật viên, thợ sửa chữa.",
		},
		{
			"ORGANIZE", "Tổ chức (Organization)",
			"Khả năng quản lý, tổ chức và lãnh đạo",
			"Người có khả năng tổ chức tốt thích lên kế hoạch, điều phối, quản lý người khác. Họ có thể làm quản lý, giám đốc, event planner, HR, project manager.",
		},
		{
			"PERSUADE", "Thuyết phục (Persuasion)",
			"Khả năng thuyết phục, bán hàng và tạo ảnh hưởng tích cực",
			"Người có khả năng thuyết phục tốt biết cách giao tiếp để thuyết phục người khác. Họ phù hợp làm sales, marketing, quan hệ công chúng, chính trị gia, luật sư.",
		},
		{
			"SCIENCE", "Khoa học (Science)",
			"Khả năng nghiên cứu, thực nghiệm và khám phá khoa học",
			"Người có khiếu khoa học thích tìm hiểu, nghiên cứu, thực nghiệm. Họ có thể trở thành nhà khoa học, bác sĩ, dược sĩ, nhà nghiên cứu, kỹ sư.",
		},
	}

	vs := s.vectorStores[collection]
	var docs []schema.Document

	for _, domain := range domains {
		// Create comprehensive document for each domain
		content := fmt.Sprintf(`ILO Domain: %s (%s)

Mô tả: %s

Đặc điểm và nghề nghiệp phù hợp: %s

Từ khóa: domain, %s, %s, career guidance, nghề nghiệp, khả năng`,
			domain.name, domain.code, domain.description, domain.characteristics, domain.code, domain.name)

		doc := schema.Document{
			PageContent: content,
			Metadata: map[string]interface{}{
				"type":         "ilo_domain",
				"domain_code":  domain.code,
				"domain_name":  domain.name,
				"indexed_at":   time.Now().Format(time.RFC3339),
				"language":     "vietnamese",
			},
		}
		docs = append(docs, doc)
	}

	_, err := vs.AddDocuments(ctx, docs)
	if err != nil {
		return 0, err
	}

	log.Printf("Ingested %d domain documents", len(docs))
	return len(docs), nil
}

func (s *LLMServiceImpl) ingestCareerMappings(ctx context.Context, collection string) (int, error) {
	// Career mappings based on domain combinations
	careerMappings := []struct {
		career, domains, description, requirements string
	}{
		{
			"Lập trình viên (Software Developer)",
			"LOGIC, DESIGN",
			"Phát triển phần mềm, ứng dụng và hệ thống máy tính",
			"Tư duy logic mạnh, khả năng giải quyết vấn đề, học hỏi công nghệ mới, kiên nhẫn debug code",
		},
		{
			"Nhà báo (Journalist)",
			"LANG, PERSUADE",
			"Thu thập, viết và truyền tải thông tin đến công chúng",
			"Khả năng viết tốt, giao tiếp hiệu quả, tò mò, có trách nhiệm với thông tin",
		},
		{
			"Kiến trúc sư (Architect)",
			"DESIGN, LOGIC, MECHANIC",
			"Thiết kế và lập kế hoạch xây dựng các công trình",
			"Tư duy không gian, sáng tạo, hiểu biết kỹ thuật, khả năng làm việc nhóm",
		},
		{
			"Bác sĩ (Doctor)",
			"SCIENCE, ORGANIZE, PERSUADE",
			"Chẩn đoán và điều trị bệnh nhân",
			"Kiến thức y khoa vững, kỹ năng giao tiếp, đồng cảm, khả năng làm việc dưới áp lực",
		},
		{
			"Giáo viên (Teacher)",
			"LANG, ORGANIZE, PERSUADE",
			"Giảng dạy và hướng dẫn học sinh",
			"Khả năng truyền đạt kiến thức, kiên nhẫn, kỹ năng quản lý lớp học, đam mê giáo dục",
		},
		{
			"Marketing Manager",
			"PERSUADE, DESIGN, ORGANIZE",
			"Phát triển và thực hiện các chiến lược marketing",
			"Hiểu thị trường, sáng tạo, khả năng phân tích dữ liệu, kỹ năng lãnh đạo",
		},
		{
			"Nhạc sĩ (Musician)",
			"MUSIC, DESIGN",
			"Sáng tác, biểu diễn và sản xuất âm nhạc",
			"Tai nhạy bén, sáng tạo, kỹ thuật chơi nhạc cụ, hiểu lý thuyết âm nhạc",
		},
		{
			"Kỹ sư cơ khí (Mechanical Engineer)",
			"MECHANIC, LOGIC, SCIENCE",
			"Thiết kế và phát triển máy móc, thiết bị",
			"Hiểu biết vật lý, toán học, khả năng thiết kế CAD, tư duy logic",
		},
	}

	vs := s.vectorStores[collection]
	var docs []schema.Document

	for _, career := range careerMappings {
		content := fmt.Sprintf(`Nghề nghiệp: %s

Các domain ILO phù hợp: %s

Mô tả công việc: %s

Yêu cầu và kỹ năng cần thiết: %s

Từ khóa: career, nghề nghiệp, %s, career guidance, %s`,
			career.career, career.domains, career.description, career.requirements, career.career, career.domains)

		doc := schema.Document{
			PageContent: content,
			Metadata: map[string]interface{}{
				"type":         "career_mapping",
				"career_name":  career.career,
				"domains":      career.domains,
				"indexed_at":   time.Now().Format(time.RFC3339),
				"language":     "vietnamese",
			},
		}
		docs = append(docs, doc)
	}

	_, err := vs.AddDocuments(ctx, docs)
	if err != nil {
		return 0, err
	}

	log.Printf("Ingested %d career mapping documents", len(docs))
	return len(docs), nil
}

func (s *LLMServiceImpl) ingestGuidanceTemplates(ctx context.Context, collection string) (int, error) {
	// Career guidance templates and advice
	templates := []struct {
		title, content, category string
	}{
		{
			"Hướng dẫn chọn ngành cho học sinh phổ thông",
			`Khi chọn ngành nghề, học sinh cần xem xét:

1. Kết quả bài test ILO để hiểu rõ khả năng và sở thích
2. Điều kiện gia đình và khả năng tài chính
3. Xu hướng thị trường lao động
4. Cơ hội phát triển nghề nghiệp
5. Môi trường làm việc mong muốn

Nên tham khảo ý kiến từ gia đình, thầy cô, và các chuyên gia hướng nghiệp.`,
			"general_guidance",
		},
		{
			"Phát triển kỹ năng dựa trên kết quả ILO",
			`Dựa trên điểm số cao nhất trong bài test ILO:

- LANG cao: Tham gia câu lạc bộ văn học, luyện nói thuyết trình, học ngoại ngữ
- LOGIC cao: Giải toán, học lập trình, tham gia Olympic tin học
- DESIGN cao: Học vẽ, thiết kế, tham gia các khóa học sáng tạo
- MUSIC cao: Học nhạc cụ, tham gia ban nhạc, học sáng tác
- MECHANIC cao: Tham gia câu lạc bộ robotics, học CAD, thực hành với máy móc
- ORGANIZE cao: Tham gia học sinh cán bộ, tổ chức sự kiện
- PERSUADE cao: Tham gia đội tuyên truyền, học bán hàng, thuyết trình
- SCIENCE cao: Tham gia câu lạc bộ khoa học, làm thí nghiệm, nghiên cứu khoa học`,
			"skill_development",
		},
		{
			"Xu hướng nghề nghiệp trong thời đại số",
			`Các nghề nghiệp hot trong kỷ nguyên số:

1. Data Scientist - Phân tích dữ liệu lớn
2. AI/ML Engineer - Phát triển trí tuệ nhân tạo
3. Cybersecurity Specialist - Bảo mật thông tin
4. Digital Marketing - Marketing số
5. UX/UI Designer - Thiết kế trải nghiệm người dùng
6. Cloud Architect - Kiến trúc đám mây
7. DevOps Engineer - Vận hành phát triển
8. Blockchain Developer - Phát triển blockchain

Tất cả đều cần kết hợp nhiều kỹ năng từ các domain ILO khác nhau.`,
			"future_trends",
		},
	}

	vs := s.vectorStores[collection]
	var docs []schema.Document

	for _, template := range templates {
		doc := schema.Document{
			PageContent: fmt.Sprintf(`%s

%s

Từ khóa: career guidance, hướng nghiệp, %s, tư vấn nghề nghiệp`,
				template.title, template.content, template.category),
			Metadata: map[string]interface{}{
				"type":       "guidance_template",
				"title":      template.title,
				"category":   template.category,
				"indexed_at": time.Now().Format(time.RFC3339),
				"language":   "vietnamese",
			},
		}
		docs = append(docs, doc)
	}

	_, err := vs.AddDocuments(ctx, docs)
	if err != nil {
		return 0, err
	}

	log.Printf("Ingested %d guidance template documents", len(docs))
	return len(docs), nil
}
