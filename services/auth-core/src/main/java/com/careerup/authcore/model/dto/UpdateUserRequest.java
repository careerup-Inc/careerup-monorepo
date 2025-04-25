package com.careerup.authcore.model.dto;

import lombok.Data;

@Data
public class UpdateUserRequest {
    private String email;
    private String firstName;
    private String lastName;
    private String hometown;
    private java.util.List<String> interests;
}
