"""Room calendar additions to the implemented API artifacts."""
def extend(schemas, op, paths, obj, ref, array, S, I, B, DATE, STAMP):
    schemas["Room"]["properties"].update(calendar_revision=I,calendar_required=B,calendar_prepared_through=S,calendar_job_error=S)
    schemas["Session"]["properties"].update(origin={"type":"string","enum":["MANUAL","CALENDAR"]},closure_source=S)
    schemas["SlotInput"]["description"] += " End may be 24:00; start may not."
    schemas["SessionCreate"]["description"]="AVAILABLE room, active slot and future start required. Entire session must fit effective published opening hours. Explicit legacy transition rooms are exempt until their first weekly publication. No overlapping OPEN sessions."
    schemas.update({
        "Hours":obj(dict(weekday={"type":"integer","minimum":0,"maximum":7},start_minute={"type":"integer","minimum":0,"maximum":1380,"multipleOf":60},end_minute={"type":"integer","minimum":60,"maximum":1440,"multipleOf":60}),["weekday","start_minute","end_minute"]),
        "ScheduleVersion":obj(dict(id=I,room_id=I,effective_from=DATE,effective_until=S,hours=array(ref("Hours")),reason=S,status=S,created_at=STAMP)),
        "CalendarException":obj(dict(id=I,room_id=I,date=DATE,mode=S,closure_type=S,reason=S,status=S,hours=array(ref("Hours")),updated_at=STAMP)),
        "CalendarSnapshot":obj(dict(room=ref("Room"),versions=array(ref("ScheduleVersion")),exceptions=array(ref("CalendarException")),slots=array(ref("TimeSlot")),sessions=array(ref("Session")))),
        "CalendarDay":obj(dict(date=DATE,source=S,version_id=I,exception_id=I,hours=array(ref("Hours")))),
        "CalendarChange":obj(dict(session_id=I,date=DATE,starts_at=STAMP,ends_at=STAMP,active_holds=I,action={"type":"string","enum":["CLOSE","REOPEN"]})),
        "CalendarPreview":obj(dict(preview_id=S,expires_at=STAMP,base_revision=I,changes=array(ref("CalendarChange")),conflicts=array(ref("CalendarChange")),creates=I,prepared_through=S)),
        "CalendarResult":obj(dict(calendar_revision=I,created=I,closed=I,reopened=I,prepared_through=S)),
        "CalendarHistory":obj(dict(kind=S,version=ref("ScheduleVersion"),exception=ref("CalendarException"))),
    })
    fields=dict(kind={"type":"string","enum":["WEEKLY","EXCEPTION","WITHDRAW","EXTEND"]},base_revision={**I,"minimum":0},effective_from=DATE,date=DATE,mode={"type":"string","enum":["CLOSED","REPLACE"]},closure_type={"type":"string","enum":["HOLIDAY","MAINTENANCE","OTHER"]},exception_id=I,hours=array(ref("Hours")),reason={"type":"string","minLength":1,"maxLength":1000},days={"type":"integer","minimum":1,"maximum":56,"default":28})
    description="WEEKLY: effective_from today for first schedule, otherwise tomorrow+, weekday 1..7. EXCEPTION: date today..56 days, mode/closure_type, weekday 0; optional active exception_id. CLOSED has no hours, REPLACE needs hours. WITHDRAW requires exception_id. EXTEND fills missing only. Half-open hourly intervals; adjacent merged, overlap rejected. Irrelevant fields rejected."
    for name,extra,required in [("CalendarMutation",{},["kind","base_revision","hours","reason"]),("CalendarConfirm",{"preview_id":S},["kind","base_revision","hours","reason","preview_id"])]:
        schemas[name]=obj({**fields,**extra},required);schemas[name].update(additionalProperties=False,description=description)
    ops="ADMIN or MANAGER"
    weekly=dict(kind="WEEKLY",base_revision="{{calendar_revision}}",effective_from="{{effective_date}}",hours=[dict(weekday=d,start_minute=420,end_minute=1260) for d in range(1,8)],reason="Local calendar demo")
    op("GET","/rooms/{id}/weekly-schedule","Read management calendar snapshot",roles=ops,output="CalendarSnapshot")
    op("GET","/rooms/{id}/calendar","Read published effective opening hours",output="CalendarDay",sequence=True,filters=("from","to"),description="Required inclusive from/to dates, maximum 56 days. No exception reasons, drafts or booking contacts.")
    for p in paths["/rooms/{id}/calendar"]["get"]["parameters"]:
        if p["name"] in ("from","to"):p.update(required=True,schema=DATE)
    op("GET","/rooms/{id}/calendar/history","Read calendar history",roles=ops,output="CalendarHistory",paged=True)
    op("POST","/rooms/{id}/calendar/preview","Preview without writing",roles=ops,output="CalendarPreview",body="CalendarMutation",example=weekly,description="Five-minute actor/payload/revision/dependency-bound preview. Affected session IDs, times and held counts are returned. Uncommitted previews become invalid after API restart. Committed requests remain replayable from database idempotency records.")
    for method,path,kind,title in [("PUT","/rooms/{id}/weekly-schedule","WEEKLY","Publish weekly schedule"),("POST","/rooms/{id}/calendar/exceptions","EXCEPTION","Create date exception"),("PATCH","/rooms/{id}/calendar/exceptions/{exceptionId}","EXCEPTION","Replace exception, retain history"),("POST","/rooms/{id}/calendar/exceptions/{exceptionId}/withdraw","WITHDRAW","Withdraw exception"),("POST","/rooms/{id}/calendar/materialize","EXTEND","Fill missing future sessions")]:
        example={**weekly,"preview_id":"{{preview_id}}"} if kind=="WEEKLY" else dict(kind=kind,base_revision="{{calendar_revision}}",hours=[],reason="Local calendar demo",preview_id="{{preview_id}}")
        if kind=="EXCEPTION":example.update(date="{{exception_date}}",mode="CLOSED",closure_type="HOLIDAY")
        if "{exceptionId}" in path:example["exception_id"]="{{exception_id}}"
        op(method,path,title,roles=ops,output="CalendarResult",body="CalendarConfirm",example=example,description="Requires matching preview and Idempotency-Key. Same key/payload replays; different payload conflicts. Room/slot/session locks recheck holds. Stale data, affected PENDING_PT/CONFIRMED bookings or audit failure roll back all writes. Manual closures/past/ongoing snapshots stay unchanged.")
        paths[path][method.lower()]["parameters"].append({"name":"Idempotency-Key","in":"header","required":True,"schema":{"type":"string","minLength":8,"maxLength":100}})
