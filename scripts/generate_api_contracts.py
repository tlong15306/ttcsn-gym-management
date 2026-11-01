"""Generate current OpenAPI/Postman artifacts; --check also compares Gin routes.

Only public demo values are included. This does not execute Postman mutations.
OpenAPI JSON follows https://spec.openapis.org/oas/v3.0.3.html.
"""
import argparse
import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
S = {"type": "string"}
I = {"type": "integer", "format": "int64"}
B = {"type": "boolean"}
DATE = {"type": "string", "format": "date"}
TIME = {"type": "string", "pattern": r"^\d{2}:\d{2}$", "example": "08:00"}
STAMP = {"type": "string", "format": "date-time"}
ROLE = {"type": "string", "enum": ["ADMIN", "MANAGER", "RECEPTIONIST", "MEMBER", "PT"]}
PASSWORD = {"type": "string", "format": "password", "minLength": 8, "description": "At least 8 Unicode characters, at most 72 UTF-8 bytes; bcrypt."}

def ref(name):
    return {"$ref": "#/components/schemas/" + name}

def array(item):
    return {"type": "array", "items": item}

def obj(fields, required=()):
    result = {"type": "object", "properties": fields}
    if required:
        result["required"] = list(required)
    return result

schemas = {
    "User": obj(dict(id=I, username=S, full_name=S, email=S, phone=S, role=ROLE, status=S, created_at=STAMP, updated_at=STAMP)),
    "Contact": obj(dict(full_name=S, email={"type":"string","format":"email"}, phone=S), ["full_name","email"]),
    "AccountCreate": obj(dict(username=S, password=PASSWORD, full_name=S, email=S, phone=S, role=ROLE), ["username","password","full_name","email","role"]),
    "AccountStatus": obj(dict(status={"type":"string","enum":["ACTIVE","LOCKED"]},reason=S), ["status","reason"]),
    "AccountRole": obj(dict(role=ROLE,reason=S), ["role","reason"]),
    "PasswordReset": obj(dict(password=PASSWORD), ["password"]),
    "LoginInput": obj(dict(username=S,password={"type":"string","format":"password"}), ["username","password"]),
    "Login": obj(dict(access_token=S,expires_in=I,token_type=S,user=ref("User"))),
    "Member": obj(dict(id=I,user_id=I,user=ref("User"),member_code=S,date_of_birth=S,gender=S,address=S,join_date=DATE,membership_expiry=DATE,status=S,expired=B)),
    "MemberCreate": obj(dict(user_id=I,account=obj(dict(username=S,password=PASSWORD,full_name=S,email=S,phone=S),["username","password","full_name","email"]),member_code=S,date_of_birth=S,gender=S,address=S,join_date=DATE,membership_expiry=DATE), ["member_code","join_date","membership_expiry"]),
    "MemberUpdate": obj(dict(full_name=S,email=S,phone=S,date_of_birth=S,gender=S,address=S), ["full_name","email"]),
    "Membership": obj(dict(membership_expiry=DATE,status={"type":"string","enum":["ACTIVE","INACTIVE"]},reason=S), ["membership_expiry","status","reason"]),
    "Room": obj(dict(id=I,name=S,room_type=S,default_capacity=I,description=S,status=S)),
    "RoomInput": obj(dict(name=S,room_type=S,default_capacity={"type":"integer","minimum":1,"maximum":10000},description=S,status={"type":"string","enum":["AVAILABLE","MAINTENANCE","INACTIVE"]},reason=S), ["name","room_type","default_capacity"]),
    "Impact": obj(dict(future_sessions=I,active_holds=I)),
    "TimeSlot": obj(dict(id=I,name=S,start_time=TIME,end_time=TIME,is_active=B)),
    "SlotInput": obj(dict(name=S,start_time=TIME,end_time=TIME,reason=S),["name","start_time","end_time"]),
    "SlotStatus": obj(dict(is_active=B,reason=S),["is_active","reason"]),
    "Session": obj(dict(id=I,room_id=I,room_name=S,time_slot_id=I,slot_name=S,session_date=DATE,starts_at=STAMP,ends_at=STAMP,capacity=I,active_holds=I,status=S,change_reason=S)),
    "SessionCreate": obj(dict(room_id=I,time_slot_id=I,session_date=DATE,capacity={"type":"integer","minimum":0,"maximum":10000,"default":0}),["room_id","time_slot_id","session_date"]),
    "SessionUpdate": obj(dict(capacity={"type":"integer","minimum":1,"maximum":10000},status={"type":"string","enum":["OPEN","CLOSED","MAINTENANCE"]},reason=S),["capacity","status","reason"]),
    "Trainer": obj(dict(id=I,user_id=I,user=ref("User"),specialty=S,experience_text=S,has_certificate=B,approval_status=S,decision_reason=S,decided_at={**STAMP,"nullable":True},rating_average={**{"type":"number","format":"float"},"nullable":True},review_count=I)),
    "TrainerInput": obj(dict(specialty=S,experience_text=S),["specialty","experience_text"]),
    "Decision": obj(dict(action={"type":"string","enum":["APPROVE","REJECT","SUSPEND"]},reason=S),["action","reason"]),
    "WorkingHour": obj(dict(id=I,day_of_week={"type":"integer","minimum":1,"maximum":7},start_time=TIME,end_time=TIME)),
    "Availability": obj(dict(hours={**array(ref("WorkingHour")),"maxItems":50}),["hours"]),
    "Absence": obj(dict(id=I,starts_at=STAMP,ends_at=STAMP,reason=S)),
    "AbsenceInput": obj(dict(starts_at=STAMP,ends_at=STAMP,reason=S),["starts_at","ends_at","reason"]),
    "CertificateUpload": obj(dict(file={"type":"string","format":"binary","description":"PDF/PNG/JPEG, decoded server-side, at most 5 MiB."}),["file"]),
    "PageMeta": obj(dict(page=I,page_size=I,total=I),["page","page_size","total"]),
    "Error": obj(dict(error=obj(dict(code=S,message=S),["code","message"]),request_id=S),["error","request_id"]),
}
schemas["MemberCreate"]["description"] = "Exactly one of positive user_id or account. Existing account must be MEMBER without a profile. New account is forced to MEMBER atomically."
schemas["RoomInput"]["description"] = "PATCH requires status and nonempty reason. Unavailable room closes future OPEN sessions; blocked by holding bookings. Capacity change affects only future snapshots."
schemas["SlotInput"]["description"] = "Same-day HH:mm, end > start. Active slots do not overlap. PATCH requires reason; used slots are immutable."
AUTH = "All authenticated roles"
OPS = "ADMIN or MANAGER"
MEM = "ADMIN, MANAGER or RECEPTIONIST"
SELFPT = "PT SELF"
paths = {}

def query(name, schema=S):
    return {"name":name,"in":"query","required":False,"schema":schema}

def op(method, path, summary, roles=AUTH, output="User", body=None, status=200, paged=False, sequence=False, filters=(), description="", example=None):
    params = [{"name":n,"in":"path","required":True,"schema":{**I,"minimum":1}} for n in re.findall(r"\{(\w+)\}",path)]
    params += [{"name":"X-Request-ID","in":"header","required":False,"schema":S}]
    if paged:
        params += [query("page",{"type":"integer","minimum":1,"maximum":1000000,"default":1}),query("page_size",{"type":"integer","minimum":1,"maximum":100,"default":20})]
    params += [query(n, STAMP if n=="available_at" else DATE if n in ("date", "from", "to") else I if n=="room_id" else B if n=="approved" else S) for n in filters]
    responses = {str(status):{"description":"Success"}}
    if status != 204:
        if output=="binary":
            responses[str(status)]["content"]={t:{"schema":{"type":"string","format":"binary"}} for t in ["application/pdf","image/png","image/jpeg"]}
        else:
            data = array(ref(output)) if paged or sequence else ref(output)
            fields={"data":data,"request_id":S}
            if paged: fields["meta"]=ref("PageMeta")
            responses[str(status)]["content"]={"application/json":{"schema":obj(fields,fields.keys())}}
    for code in [401,403,404,409,422,500]:
        responses[str(code)]={"description":"Authentication, permission, scope, state, validation or internal error.","content":{"application/json":{"schema":ref("Error")}}}
    value={"summary":summary,"operationId":method.lower()+re.sub(r"[^a-zA-Z0-9]","_",path),"tags":[path.split('/')[1]],"parameters":params,"responses":responses,"x-roles":roles,"description":roles+". "+description}
    if roles=="Public":value["security"]=[]
    if body:
        media="multipart/form-data" if body=="CertificateUpload" else "application/json"
        value["requestBody"]={"required":True,"content":{media:{"schema":ref(body)}}}
    if example is not None:value["x-demo-body"]=example
    paths.setdefault(path,{})[method.lower()]=value

contact=dict(full_name="Demo Admin CRUD",email="crud@example.local",phone="")
op("POST","/auth/login","Login",roles="Public",output="Login",body="LoginInput",example=dict(username="admin.demo",password="DemoPass123!"))
op("POST","/auth/logout","Revoke all account sessions",status=204)
op("GET","/me","Read own account")
op("PATCH","/me","Update own contact",body="Contact",example=contact)
op("GET","/users","List accounts",roles="ADMIN",paged=True,filters=("q","role","status"))
op("POST","/users","Create account",roles="ADMIN",body="AccountCreate",status=201,example=dict(**contact,username="crud.demo",password="DemoPass123!",role="MEMBER"))
op("GET","/users/{id}","Account detail",roles="ADMIN")
op("PATCH","/users/{id}","Edit account contact",roles="ADMIN",body="Contact",example=contact)
op("PATCH","/users/{id}/status","Lock or unlock",roles="ADMIN",body="AccountStatus",example=dict(status="LOCKED",reason="Local demo"))
op("PATCH","/users/{id}/role","Change role",roles="ADMIN",body="AccountRole",example=dict(role="RECEPTIONIST",reason="Local demo"))
op("POST","/users/{id}/reset-password","Reset password",roles="ADMIN",body="PasswordReset",example=dict(password="FreshPass123!"))
op("GET","/members","List scoped members",roles=MEM+"; MEMBER SELF; PT ASSIGNED",output="Member",paged=True,filters=("q","status"))
op("GET","/members/eligible-accounts","Unlinked MEMBER accounts",roles=MEM,paged=True,filters=("q",))
op("POST","/members","Create or link member",roles=MEM,output="Member",body="MemberCreate",status=201,example=dict(user_id="{{user_id}}",member_code="HV-CRUD-DEMO",join_date="2026-10-11",membership_expiry="2027-10-11"))
op("GET","/members/{id}","Scoped member detail",roles=MEM+"; MEMBER SELF; PT ASSIGNED",output="Member")
op("PATCH","/members/{id}","Edit member profile",roles=MEM,output="Member",body="MemberUpdate",example=dict(**contact,date_of_birth="",gender="",address=""))
op("PATCH","/members/{id}/membership","Update membership",roles=MEM,output="Member",body="Membership",example=dict(membership_expiry="2027-10-11",status="ACTIVE",reason="Local demo"))
op("GET","/rooms","List rooms",output="Room",paged=True,filters=("q","status"))
room=dict(name="Gym CRUD demo",room_type="Gym",default_capacity=20,status="AVAILABLE")
op("POST","/rooms","Create room",roles=OPS,output="Room",body="RoomInput",status=201,example=room)
op("GET","/rooms/{id}","Room detail",output="Room")
op("PATCH","/rooms/{id}","Update room",roles=OPS,output="Room",body="RoomInput",example=dict(**room,reason="Local demo"))
op("GET","/rooms/{id}/impact","Preview future session impact",roles=OPS,output="Impact")
op("GET","/time-slots","List slots",output="TimeSlot",paged=True,filters=("q","status"))
slot=dict(name="Morning demo",start_time="08:00",end_time="09:00")
op("POST","/time-slots","Create slot",roles=OPS,output="TimeSlot",body="SlotInput",status=201,example=slot)
op("GET","/time-slots/{id}","Slot detail",output="TimeSlot")
op("PATCH","/time-slots/{id}","Edit unused slot",roles=OPS,output="TimeSlot",body="SlotInput",example=dict(**slot,reason="Local demo"))
op("PATCH","/time-slots/{id}/status","Activate or retire slot",roles=OPS,output="TimeSlot",body="SlotStatus",example=dict(is_active=False,reason="Local demo"))
op("GET","/room-sessions","List sessions",output="Session",paged=True,filters=("q","status","date","room_id"))
op("POST","/room-sessions","Create snapshot in Vietnam timezone",roles=OPS,output="Session",body="SessionCreate",status=201,example=dict(room_id="{{room_id}}",time_slot_id="{{slot_id}}",session_date="2026-10-25",capacity=0))
op("GET","/room-sessions/{id}","Session detail",output="Session")
op("PATCH","/room-sessions/{id}","Change capacity or state safely",roles=OPS,output="Session",body="SessionUpdate",example=dict(capacity=20,status="OPEN",reason="Local demo"))
op("GET","/pts","List permitted PT profiles",output="Trainer",paged=True,filters=("q","status","approved","available_at"),description="Operators see all; others see active approved PT and their own profile. available_at checks a single instant, not an entire booking interval. Internal decision details are withheld from other roles.")
op("GET","/pts/me","Own PT profile",roles=SELFPT,output="Trainer")
op("GET","/pts/{id}","Scoped PT detail",output="Trainer")
op("PUT","/pts/me/profile","Save DRAFT or REJECTED profile",roles=SELFPT,output="Trainer",body="TrainerInput",example=dict(specialty="Gym demo",experience_text="Synthetic local QA profile"))
op("POST","/pts/me/submit","Submit complete profile to PENDING",roles=SELFPT,output="Trainer")
op("POST","/pts/me/certificates","Upload private certificate",roles=SELFPT,output="Trainer",body="CertificateUpload")
op("GET","/pts/{id}/certificate","Download private certificate",roles=OPS+" or PT SELF",output="binary")
op("POST","/pts/{id}/decision","Decide profile",roles=OPS,output="Trainer",body="Decision",example=dict(action="APPROVE",reason="Synthetic local QA decision"))
for prefix, roles in [("/pts/me",SELFPT),("/pts/{id}",OPS+" or PT SELF")]:
    op("GET",prefix+"/availability","Read weekly hours",roles=roles,output="WorkingHour",sequence=True)
    op("PUT",prefix+"/availability","Replace weekly hours atomically",roles=roles,output="WorkingHour",sequence=True,body="Availability",example=dict(hours=[dict(day_of_week=1,start_time="08:00",end_time="17:00")]))
    op("GET",prefix+"/unavailability","List private absences",roles=roles,output="Absence",paged=True)
    op("POST",prefix+"/unavailability","Add absence",roles=roles,output="Absence",body="AbsenceInput",status=201,example=dict(starts_at="2026-12-01T12:00:00+07:00",ends_at="2026-12-01T13:00:00+07:00",reason="Synthetic local QA"))
    op("DELETE",prefix+"/unavailability/{absenceId}","Withdraw absence, retain audit",roles=roles,status=204)

from calendar_contract import extend
extend(schemas,op,paths,obj,ref,array,S,I,B,DATE,STAMP)
from booking_contract import extend as extend_booking
extend_booking(schemas,op,paths,obj,ref,array,S,I,B,DATE,STAMP)
from review_contract import extend as extend_review
extend_review(schemas,op,paths,obj,ref,array,S,I,B,DATE,STAMP)
from report_contract import extend as extend_report
extend_report(schemas,op,paths,obj,ref,array,S,I,B,DATE,STAMP)
from discovery_contract import extend as extend_discovery
extend_discovery(schemas,op,paths,obj,ref,array,S,I,B,DATE,STAMP)

document={"openapi":"3.0.3","info":{"title":"TTCSN GYM Admin CRUD","version":"1.1.0","description":"Implemented Identity, Members, Facilities, Booking, Attendance, Review and reports. Local demo only; no production release."},"servers":[{"url":"http://127.0.0.1:8080/api/v1"}],"security":[{"bearerAuth":[]}],"paths":paths,"components":{"securitySchemes":{"bearerAuth":{"type":"http","scheme":"bearer","bearerFormat":"JWT"}},"schemas":schemas}}

def collection():
    document["info"].update(version="1.6.0",description="Implemented Identity, Members, Facilities, room calendar, room covers, PT, Booking, Attendance, Review, scoped reports and public discovery. Local demo only.")
    groups={}
    for path, methods in paths.items():
        for method, operation in methods.items():
            entity=path.split('/')[1]
            idvar={"users":"user_id","members":"member_id","rooms":"room_id","time-slots":"slot_id","room-sessions":"session_id","pts":"pt_id","bookings":"booking_id"}.get(entity,"id")
            raw="{{base_url}}"+path.replace("{id}","{{"+idvar+"}}").replace("{absenceId}","{{absence_id}}").replace("{exceptionId}","{{exception_id}}")
            if path.endswith("/calendar"):raw+="?from={{effective_date}}&to={{effective_date}}"
            request={"method":method.upper(),"header":[],"url":{"raw":raw,"host":["{{base_url}}"],"path":raw.split("}}",1)[1].split('?',1)[0].lstrip('/').split('/')},"description":operation["description"]}
            if path.endswith("/calendar"):
                request["url"]["query"]=[{"key":"from","value":"{{effective_date}}"},{"key":"to","value":"{{effective_date}}"}]
            if path == "/reports/operations":
                request["url"]["raw"] += "?from={{from_date}}&to={{to_date}}&room_id={{room_id}}"
                request["url"]["query"]=[{"key":"from","value":"{{from_date}}"},{"key":"to","value":"{{to_date}}"},{"key":"room_id","value":"{{room_id}}"}]
            if path == "/reports/pts/{id}":
                request["url"]["raw"] += "?from={{from_date}}&to={{to_date}}"
                request["url"]["query"]=[{"key":"from","value":"{{from_date}}"},{"key":"to","value":"{{to_date}}"}]
            if operation.get("security")==[]:request["auth"]={"type":"noauth"}
            if "requestBody" in operation:
                if path.endswith("/certificates") or path.endswith('/cover-image'):
                    request["body"]={"mode":"formdata","formdata":[{"key":"file","type":"file","src":[]}]}
                    if path.endswith('/cover-image'): request['body']['formdata'] += [{'key':'alt_text','type':'text','value':'Ảnh phòng tập'},{'key':'is_demo','type':'text','value':'false'}]
                else:
                    request["header"]=[{"key":"Content-Type","value":"application/json"}]
                    body=json.dumps(operation.get("x-demo-body",{}),ensure_ascii=False,indent=2)
                    body=re.sub(r'"(\{\{(?:user_id|room_id|slot_id|session_id|session_id_2|calendar_revision|exception_id)\}\})"',r'\1',body)
                    request["body"]={"mode":"raw","raw":body,"options":{"raw":{"language":"json"}}}
            item={"name":method.upper()+" "+path,"request":request,"response":[]}
            if path=="/auth/login":
                item["event"]=[{"listen":"test","script":{"type":"text/javascript","exec":["if (pm.response.code === 200) pm.collectionVariables.set('access_token', pm.response.json().data.access_token);"]}}]
            if path.endswith("/calendar/preview"):
                item["event"]=[{"listen":"test","script":{"type":"text/javascript","exec":["if (pm.response.code === 200) { pm.collectionVariables.set('preview_id', pm.response.json().data.preview_id); pm.collectionVariables.set('calendar_key', pm.variables.replaceIn('{{$guid}}')); }"]}}]
            if path.endswith("/weekly-schedule") and method=="get":
                item["event"]=[{"listen":"test","script":{"type":"text/javascript","exec":["if (pm.response.code === 200) pm.collectionVariables.set('calendar_revision', pm.response.json().data.room.calendar_revision);"]}}]
            if any(p["name"]=="Idempotency-Key" for p in operation["parameters"]):request["header"].append({"key":"Idempotency-Key","value":"{{booking_key}}" if path in ['/bookings','/members/{id}/bookings','/bookings/batch','/members/{id}/bookings/batch'] else "{{calendar_key}}"})
            if path == '/booking-options/sessions': request['url']['raw'] += '?room_id={{room_id}}&date={{booking_date}}'
            if path == '/booking-options/pts': request['url']['raw'] += '?room_session_id={{session_id}}'
            groups.setdefault(entity,[]).append(item)
    return {"info":{"name":"GYM Admin CRUD - implemented endpoints","schema":"https://schema.getpostman.com/json/collection/v2.1.0/collection.json","description":"Use only local demo. Select role by changing login credentials. Set IDs after creation. Mutations include reset/logout/retire; do not blindly Run All. Collection generated and route checked; not executed in Postman."},"auth":{"type":"bearer","bearer":[{"key":"token","value":"{{access_token}}","type":"string"}]},"variable":[{"key":k,"value":v,"type":"string"} for k,v in {"base_url":"http://127.0.0.1:8080/api/v1","access_token":"","user_id":"1","member_id":"1","room_id":"1","slot_id":"1","session_id":"1","session_id_2":"2","pt_id":"1","absence_id":"1"}.items()],"item":[{"name":k,"item":v} for k,v in groups.items()]}

def verify_routes():
    source=(ROOT/'server/internal/routes/v1/router.go').read_text(encoding='utf-8')
    actual=set()
    for group,method,path in re.findall(r'(api|accounts|authenticated|public)\.(GET|POST|PATCH|PUT|DELETE)\("([^"]*)"',source):
        path=({"accounts":"/users","public":"/public"}.get(group,""))+path
        actual.add((method,re.sub(r':(\w+)',r'{\1}',path)))
    for method,suffix in re.findall(r'authenticated\.(GET|POST|PATCH|PUT|DELETE)\(prefix\+"([^"]+)"',source):
        for prefix in ["/pts/me","/pts/{id}"]:
            actual.add((method,prefix+re.sub(r':(\w+)',r'{\1}',suffix)))
    expected={(method.upper(),path) for path,methods in paths.items() for method in methods}
    assert actual==expected, f"Route mismatch missing={actual-expected}, extra={expected-actual}"
    def refs(value):
        if isinstance(value,dict):
            if '$ref' in value: assert value['$ref'].split('/')[-1] in schemas
            for nested in value.values():refs(nested)
        elif isinstance(value,list):
            for nested in value:refs(nested)
    refs(document)
    return len(expected)

if __name__=='__main__':
    parser=argparse.ArgumentParser();parser.add_argument('--check',action='store_true');args=parser.parse_args()
    count=verify_routes()
    document["info"].update(version="1.4.0",description="Implemented Identity, Members, Facilities, room calendar, room covers, PT, Booking, Attendance and Review. Reports remain planned. Local demo only.")
    for name,value in [('openapi_admin_crud.json',document),('GYM_Admin_CRUD.postman_collection.json',collection())]:
        if name.endswith('postman_collection.json'):
            value['variable'] += [{'key':k,'value':v,'type':'string'} for k,v in {'calendar_revision':'0','preview_id':'','calendar_key':'','effective_date':'','exception_date':'','exception_id':'1','booking_key':'','booking_date':'','booking_id':'1','from_date':'','to_date':''}.items()]
        content=json.dumps(value,ensure_ascii=False,indent=2)+'\n';target=ROOT/'docs'/name
        if args.check:assert target.read_text(encoding='utf-8')==content, f"Stale artifact: {name}"
        else:target.write_text(content,encoding='utf-8')
    print(f'PASS: {count} implemented API operations, schema references and current artifacts')
