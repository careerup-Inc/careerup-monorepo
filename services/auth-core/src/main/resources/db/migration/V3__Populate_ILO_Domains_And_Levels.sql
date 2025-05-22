
-- Insert the 5 domains
INSERT INTO ilo_domains (code, name, description) VALUES 
('LANG', 'Ngôn ngữ', 'Khả năng hiểu, thể hiện ý tưởng, và sử dụng ngôn ngữ, bao gồm kỹ năng đọc, viết, và giao tiếp bằng lời nói'),
('LOGIC', 'Phân tích - lôgic', 'Khả năng suy luận, tư duy logic, phân tích thông tin và giải quyết vấn đề'),
('DESIGN', 'Hình học - màu sắc - thiết kế', 'Khả năng nhận thức không gian, màu sắc, và thẩm mỹ, bao gồm sáng tạo nghệ thuật và thiết kế'),
('PEOPLE', 'Làm việc với con người', 'Khả năng hiểu, đồng cảm, và tương tác hiệu quả với người khác, phát triển mối quan hệ và làm việc nhóm'),
('MECH', 'Thể chất - cơ khí', 'Khả năng vận động, khéo léo, và hiểu biết về cơ chế vận hành của các hệ thống và máy móc');

-- Insert evaluation levels
INSERT INTO ilo_levels (min_percent, max_percent, level_name, suggestion) VALUES 
(80, 100, 'Rất mạnh', 'Đây là điểm mạnh nổi bật của bạn, nên xem xét các ngành nghề liên quan đến lĩnh vực này'),
(60, 79, 'Mạnh', 'Bạn có năng lực tốt trong lĩnh vực này, có thể phát triển thêm để trở thành thế mạnh'),
(40, 59, 'Trung bình', 'Bạn có khả năng trung bình trong lĩnh vực này, cần cân nhắc kết hợp với các lĩnh vực mạnh hơn'),
(0, 39, 'Yếu', 'Đây không phải là thế mạnh của bạn, nên tránh chọn nghề nghiệp đòi hỏi cao về lĩnh vực này');

-- Insert career suggestions for LANG domain
INSERT INTO ilo_career_map (domain_id, career_field, description, priority) VALUES 
((SELECT id FROM ilo_domains WHERE code = 'LANG'), 'Báo chí và Truyền thông', 'Phóng viên, biên tập viên, người dẫn chương trình', 100),
((SELECT id FROM ilo_domains WHERE code = 'LANG'), 'Giáo dục và Đào tạo', 'Giáo viên ngôn ngữ, giảng viên, nhà giáo dục', 90),
((SELECT id FROM ilo_domains WHERE code = 'LANG'), 'Biên-Phiên dịch', 'Biên dịch viên, phiên dịch viên, chuyên gia ngôn ngữ', 85),
((SELECT id FROM ilo_domains WHERE code = 'LANG'), 'Marketing và Quảng cáo', 'Copywriter, content creator, chuyên gia PR', 80),
((SELECT id FROM ilo_domains WHERE code = 'LANG'), 'Xuất bản và Sáng tạo nội dung', 'Nhà văn, biên tập sách, nhà phê bình văn học', 75);

-- Insert career suggestions for LOGIC domain
INSERT INTO ilo_career_map (domain_id, career_field, description, priority) VALUES 
((SELECT id FROM ilo_domains WHERE code = 'LOGIC'), 'Công nghệ thông tin', 'Lập trình viên, kỹ sư phần mềm, chuyên gia phân tích dữ liệu', 100),
((SELECT id FROM ilo_domains WHERE code = 'LOGIC'), 'Tài chính và Kế toán', 'Chuyên gia tài chính, kế toán viên, kiểm toán viên', 95),
((SELECT id FROM ilo_domains WHERE code = 'LOGIC'), 'Nghiên cứu khoa học', 'Nhà nghiên cứu, chuyên gia R&D, nhà khoa học dữ liệu', 90),
((SELECT id FROM ilo_domains WHERE code = 'LOGIC'), 'Quản lý dự án', 'Quản lý dự án, nhà phân tích kinh doanh, quản lý sản phẩm', 85),
((SELECT id FROM ilo_domains WHERE code = 'LOGIC'), 'Tư vấn chiến lược', 'Tư vấn viên, nhà hoạch định chiến lược, chuyên gia tối ưu hóa', 80);

-- Insert career suggestions for DESIGN domain
INSERT INTO ilo_career_map (domain_id, career_field, description, priority) VALUES 
((SELECT id FROM ilo_domains WHERE code = 'DESIGN'), 'Thiết kế đồ họa', 'Nhà thiết kế đồ họa, UI/UX designer, hoạ sĩ kỹ thuật số', 100),
((SELECT id FROM ilo_domains WHERE code = 'DESIGN'), 'Kiến trúc và Thiết kế nội thất', 'Kiến trúc sư, nhà thiết kế nội thất, nhà quy hoạch đô thị', 95),
((SELECT id FROM ilo_domains WHERE code = 'DESIGN'), 'Nghệ thuật thị giác', 'Hoạ sĩ, nhà làm phim, nhà sáng tạo đa phương tiện', 90),
((SELECT id FROM ilo_domains WHERE code = 'DESIGN'), 'Thiết kế sản phẩm', 'Nhà thiết kế sản phẩm, nhà thiết kế thời trang, nhà sáng tạo đồ thủ công', 85),
((SELECT id FROM ilo_domains WHERE code = 'DESIGN'), 'Quảng cáo và Thương hiệu', 'Nhà thiết kế thương hiệu, chuyên gia sáng tạo quảng cáo', 80);

-- Insert career suggestions for PEOPLE domain
INSERT INTO ilo_career_map (domain_id, career_field, description, priority) VALUES 
((SELECT id FROM ilo_domains WHERE code = 'PEOPLE'), 'Quản lý nhân sự', 'Giám đốc nhân sự, chuyên gia tuyển dụng, chuyên gia đào tạo', 100),
((SELECT id FROM ilo_domains WHERE code = 'PEOPLE'), 'Tư vấn và Trị liệu', 'Nhà tâm lý học, nhà tư vấn, nhà trị liệu', 95),
((SELECT id FROM ilo_domains WHERE code = 'PEOPLE'), 'Bán hàng và Kinh doanh', 'Chuyên viên kinh doanh, đại diện bán hàng, chuyên gia chăm sóc khách hàng', 90),
((SELECT id FROM ilo_domains WHERE code = 'PEOPLE'), 'Giáo dục và Đào tạo', 'Giáo viên, huấn luyện viên, nhà tư vấn giáo dục', 85),
((SELECT id FROM ilo_domains WHERE code = 'PEOPLE'), 'Dịch vụ y tế và Chăm sóc', 'Y tá, nhân viên xã hội, nhà trị liệu vật lý', 80);

-- Insert career suggestions for MECH domain
INSERT INTO ilo_career_map (domain_id, career_field, description, priority) VALUES 
((SELECT id FROM ilo_domains WHERE code = 'MECH'), 'Kỹ thuật cơ khí', 'Kỹ sư cơ khí, kỹ thuật viên sản xuất, thợ máy', 100),
((SELECT id FROM ilo_domains WHERE code = 'MECH'), 'Xây dựng và Kiến trúc', 'Kỹ sư xây dựng, giám sát công trình, thợ xây dựng chuyên nghiệp', 95),
((SELECT id FROM ilo_domains WHERE code = 'MECH'), 'Vận động và Thể thao', 'Huấn luyện viên thể thao, vận động viên chuyên nghiệp, nhà vật lý trị liệu thể thao', 90),
((SELECT id FROM ilo_domains WHERE code = 'MECH'), 'Điện tử và Tự động hóa', 'Kỹ sư điện tử, chuyên gia tự động hóa, kỹ thuật viên robot', 85),
((SELECT id FROM ilo_domains WHERE code = 'MECH'), 'Nông nghiệp và Quản lý tài nguyên', 'Kỹ sư nông nghiệp, chuyên gia lâm nghiệp, chuyên gia bảo tồn môi trường', 80);
