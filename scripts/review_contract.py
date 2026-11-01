def extend(schemas, op, paths, obj, ref, array, S, I, B, DATE, STAMP):
    schemas['Review'] = obj(dict(
        rating={"type": "integer", "minimum": 1, "maximum": 5},
        comment={**S, "maxLength": 1000},
        created_at=STAMP,
        reviewer_label={**S, "enum": ["Hội viên"]},
    ), ["rating", "comment", "created_at", "reviewer_label"])
    schemas['ReviewInput'] = obj(dict(
        rating={"type": "integer", "minimum": 1, "maximum": 5},
        comment={**S, "maxLength": 1000},
    ), ["rating"])
    schemas['ReviewInput']['additionalProperties'] = False

    op("POST", "/bookings/{id}/review", "Create one eligible PT review", roles="MEMBER SELF", output="Review", body="ReviewInput", status=201,
       example=dict(rating=5, comment="Buổi tập rất tốt."),
       description="Only a member may review their own COMPLETED booking after its end, when an assigned PT and attendance records for both participants exist. One review per booking; no edit or delete operation.")
    op("GET", "/bookings/{id}/review", "Read booking review", output="Review",
       description="Member SELF, assigned PT, and facility roles may read. data is null when no review exists; out-of-scope bookings return 404.")
    paths["/bookings/{id}/review"]["get"]["responses"]["200"]["content"]["application/json"]["schema"]["properties"]["data"] = {"allOf": [ref("Review")], "nullable": True}
    op("GET", "/pts/{id}/reviews", "List privacy-safe PT reviews", output="Review", paged=True,
       roles="MEMBER: active approved PT public reviews; PT SELF; RECEPTIONIST, MANAGER, ADMIN ALL",
       description="Returns rating, optional comment, creation date and the generic reviewer label only. Never returns member/user/booking identifiers or contact information.")
