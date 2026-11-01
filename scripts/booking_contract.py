def extend(schemas,op,paths,obj,ref,array,S,I,B,DATE,STAMP):
    schemas['Room']['properties'].update(has_cover=B,cover_alt_text=S,cover_version=I,cover_is_demo=B)
    schemas['Booking']=obj(dict(has_attendance=B,id=I,member_id=I,member_name=S,room_session_id=I,room_id=I,room_name=S,pt_id=I,pt_name=S,status={'type':'string','enum':['PENDING_PT','CONFIRMED','REJECTED','EXPIRED','CANCELLED','COMPLETED','NO_SHOW']},starts_at=STAMP,ends_at=STAMP,confirmed_at={**STAMP,'nullable':True},respond_by={**STAMP,'nullable':True},created_at=STAMP,cancellation_reason=S,cancellation_source={'type':'string','enum':['MEMBER','PT','FACILITY','UNKNOWN'],'nullable':True},cancelled_actor_role={'type':'string','enum':['ADMIN','MANAGER','RECEPTIONIST','MEMBER','PT'],'nullable':True},server_time=STAMP))
    schemas['BookingInput']=obj(dict(room_session_id=I,pt_id=I),['room_session_id'])
    schemas['BookingBatchInput']=obj(dict(room_session_ids={'type':'array','items':I,'minItems':1,'maxItems':24,'uniqueItems':True},pt_id=I),['room_session_ids'])
    schemas['BookingAction']=obj(dict(action={'type':'string','enum':['ACCEPT','REJECT']},reason=S))
    schemas['BookingCancel']=obj(dict(reason=S))
    schemas['SessionOption']={'allOf':[ref('Session'),obj(dict(remaining=I,unavailable_reason=S))]}
    schemas['PTOption']=obj(dict(id=I,full_name=S,specialty=S,rating_average={"type":"number","format":"float","nullable":True},review_count=I))
    schemas['RoomCoverUpload']=obj(dict(file={'type':'string','format':'binary','description':'JPEG/PNG <=5 MiB, <=4096 px per dimension, <=16 million pixels; re-encoded JPEG.'},alt_text={**S,'maxLength':300},is_demo=B),['file','alt_text'])
    op('GET','/booking-options/sessions','Eligible room sessions',roles='MEMBER, ADMIN, RECEPTIONIST',output='SessionOption',paged=True,filters=('room_id','date'))
    op('GET','/booking-options/pts','PT available for entire session',roles='MEMBER, ADMIN, RECEPTIONIST',output='PTOption',paged=True,filters=('room_session_id',))
    for path,roles in [('/bookings','MEMBER'),('/members/{id}/bookings','ADMIN, RECEPTIONIST')]:
        op('POST',path,'Create booking atomically',roles=roles,output='Booking',body='BookingInput',status=201,example=dict(room_session_id='{{session_id}}',pt_id=0),description='Required Idempotency-Key 8-100 characters. Same actor/key/payload replays original creation; changed payload is 409. Pending PT holds capacity until rejection/cancellation/expiry. Response deadline min(now+30 minutes, starts_at).')
        paths[path]['post']['parameters'].append(dict(name='Idempotency-Key',**{'in':'header'},required=True,schema={**S,'minLength':8,'maxLength':100}))
    for path,roles in [('/bookings/batch','MEMBER'),('/members/{id}/bookings/batch','ADMIN, RECEPTIONIST')]:
        op('POST',path,'Create contiguous bookings atomically',roles=roles,output='Booking',body='BookingBatchInput',status=201,sequence=True,example=dict(room_session_ids=['{{session_id}}','{{session_id_2}}'],pt_id=0),description='Required Idempotency-Key 8-100 characters. The server orders the selected sessions and requires one room, one date and immediately consecutive times. All selected bookings commit together or none commit. Same actor/key/payload replays the original ordered array; changed payload is 409. Invalid or non-consecutive selections return 422.')
        paths[path]['post']['parameters'].append(dict(name='Idempotency-Key',**{'in':'header'},required=True,schema={**S,'minLength':8,'maxLength':100}))
    op('GET','/bookings','Scoped booking list',output='Booking',paged=True,filters=('status','from','to','room_id','pt','q','history'),description='Member sees own bookings; PT assigned bookings; operators see all. No contact details.')
    op('GET','/bookings/{id}','Scoped booking detail',output='Booking')
    op('POST','/bookings/{id}/pt-response','Accept or reject pending booking',roles='Assigned PT',output='Booking',body='BookingAction',example=dict(action='ACCEPT',reason=''))
    op('POST','/bookings/{id}/cancel','Cancel booking',output='Booking',body='BookingCancel',example=dict(reason='Changed schedule'),description='Member self cancellation allowed at least two hours before start. Operators and assigned PT require reason. Pending PT should reject instead of cancel. Terminal transitions return 409.')
    op('POST','/rooms/{id}/cover-image','Upload or replace room cover',roles='ADMIN, MANAGER',output='Room',body='RoomCoverUpload')
    paths['/rooms/{id}/cover-image']['post']['requestBody']['content']={'multipart/form-data':{'schema':ref('RoomCoverUpload')}}
    op('DELETE','/rooms/{id}/cover-image','Delete room cover',roles='ADMIN, MANAGER',output='Room')
    op('GET','/rooms/{id}/cover-image','Read authenticated cover',output='binary',filters=('size',),description='size card (480px) or detail (1600px). JPEG with ETag and private no-cache. 304 for matching If-None-Match.')
    paths['/rooms/{id}/cover-image']['get']['responses']['200']['content']={'image/jpeg':{'schema':{'type':'string','format':'binary'}}}
    paths['/rooms/{id}/cover-image']['get']['responses']['304']={'description':'Unchanged image'}

    schemas['Attendance']=obj(dict(id=I,booking_id=I,participant_type={'type':'string','enum':['MEMBER','PT']},participant_user_id=I,checked_in_at=STAMP,recorded_by_user_id=I,recorder_name=S,override_reason=S))
    schemas['AttendanceInput']=obj(dict(participant_type={'type':'string','enum':['MEMBER','PT']},override_reason={**S,'maxLength':1000}),['participant_type'])
    schemas['AttendanceParticipant']=obj(dict(participant_type=S,name=S,attendance={'allOf':[ref('Attendance')],'nullable':True}))
    schemas['AttendanceSummary']=obj(dict(participants=array('AttendanceParticipant'),server_time=STAMP))
    op('GET','/bookings/{id}/attendance','Read actual participant attendance',output='AttendanceSummary',description='Member SELF, PT ASSIGNED, operators ALL. Missing attendance is null; self-training has no PT participant.')
    op('POST','/bookings/{id}/attendance','Record participant attendance',roles='ADMIN, MANAGER, RECEPTIONIST',output='Attendance',body='AttendanceInput',status=201,example=dict(participant_type='MEMBER',override_reason=''),description='Server infers participant and recorder. CONFIRMED only, starts_at-15min <= DB time < ends_at. Early override requires reason. First record 201, replay 200 preserves original record without extra audit; invalid window/participant 422. No additions after end; no normal cancellation after attendance.')
    paths['/bookings/{id}/attendance']['post']['responses']['200']=paths['/bookings/{id}/attendance']['post']['responses']['201']
