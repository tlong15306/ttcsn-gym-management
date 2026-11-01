def extend(schemas, op, paths, obj, ref, array, S, I, B, DATE, STAMP):
    schemas['PublicRoom'] = obj(dict(id=I,name=S,room_type=S,default_capacity=I,description=S,status={'type':'string','enum':['AVAILABLE','MAINTENANCE']},has_cover=B,cover_alt_text=S,cover_version=I,cover_is_demo=B))
    schemas['PublicTrainer'] = obj(dict(id=I,full_name=S,specialty=S,experience_text=S,rating_average={'type':'number','nullable':True},review_count=I))
    schemas['PublicSession'] = obj(dict(id=I,room_id=I,session_date=DATE,starts_at=STAMP,ends_at=STAMP,capacity=I,remaining=I,status={'type':'string','enum':['OPEN','CLOSED','MAINTENANCE']}))
    op('GET','/public/rooms','Browse public rooms',roles='Public',output='PublicRoom',paged=True,filters=('q','room_type'),description='AVAILABLE/MAINTENANCE only. Search at most 120 characters; explicit public projection, no management metadata.')
    op('GET','/public/rooms/{id}','Read a public room',roles='Public',output='PublicRoom')
    op('GET','/public/rooms/{id}/cover-image','Read public room cover',roles='Public',output='binary',description='Only visible room covers; private trainer certificates remain authenticated. ETag is revalidated, including room visibility.')
    cover=paths['/public/rooms/{id}/cover-image']['get']
    cover['parameters'] += [{'name':'size','in':'query','required':False,'schema':{'type':'string','enum':['card','detail'],'default':'card'}},{'name':'If-None-Match','in':'header','required':False,'schema':S}]
    cover['responses']['200']['content']={'image/jpeg':{'schema':{'type':'string','format':'binary'}}}
    cover['responses']['304']={'description':'Current public cover unchanged.'}
    op('GET','/public/rooms/{id}/sessions','Read public session availability',roles='Public',output='PublicSession',paged=True,filters=('from','to'),description='Required YYYY-MM-DD range, at most seven inclusive days. Future-ending, published/legacy sessions only; aggregate remaining seats, no booking identities or personalized eligibility. No holds are made.')
    for parameter in paths['/public/rooms/{id}/sessions']['get']['parameters']:
        if parameter['name'] in ('from','to'): parameter['required']=True
    op('GET','/public/trainers','Browse approved trainers',roles='Public',output='PublicTrainer',paged=True,filters=('q',),description='APPROVED trainers with active PT accounts only. No username, contact details, certificate, absence or decision data.')
