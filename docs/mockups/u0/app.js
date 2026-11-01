(() => {
  'use strict';
  const $ = selector => document.querySelector(selector);
  const $$ = selector => [...document.querySelectorAll(selector)];
  const photoRoot = '../../reports/assets/booking-portals/photos/';
  const rooms = [
    {id:'strength',name:'Strength Arena',type:'Tập sức mạnh',capacity:18,photo:'strength-demo.png',description:'Giàn tạ, ghế tập và khoảng không cho những buổi tập tập trung.'},
    {id:'cardio',name:'Cardio Studio',type:'Cardio & sức bền',capacity:16,photo:'cardio-demo.png',description:'Tìm nhịp vận động của bạn trong không gian sáng và thoáng.'},
    {id:'pt',name:'PT Studio',type:'Huấn luyện cá nhân',capacity:6,photo:'pt-demo.png',description:'Không gian riêng cho buổi tập có người hướng dẫn đồng hành.'},
  ];
  const trainers = [
    {id:'strength',name:'HLV Sức mạnh',specialty:'Strength & Conditioning',initials:'SM'},
    {id:'cardio',name:'HLV Cardio',specialty:'Cardio & Endurance',initials:'CD'},
    {id:'mobility',name:'HLV Mobility',specialty:'Mobility & Recovery',initials:'MB'},
  ];
  const weekdays = ['Thứ hai','Thứ ba','Thứ tư','Thứ năm','Thứ sáu','Thứ bảy','Chủ nhật'];
  const shortDays = ['T2','T3','T4','T5','T6','T7','CN'];
  const firstHour = 6, lastHour = 21;
  const state = {
    persona:'guest', screen:'home', room:'strength', filter:'all', query:'', scenario:'normal',
    week:0, activeDay:2, view:'week', mode:matchMedia('(max-width:650px)').matches?'view':'select',
    selection:{day:2,start:7,end:9}, training:'self', trainer:null,
  };
  let drag = null, tapAnchor = null, draft = null, toastTimer;
  const room = () => rooms.find(item => item.id === state.room);
  const pad = number => String(number).padStart(2,'0');
  const clock = hour => `${pad(hour)}:00`;
  function dateFor(day) { return new Date(Date.UTC(2026,10,2 + state.week*7 + day)); }
  function dateLabel(day, full = false) {
    const date = dateFor(day);
    return `${weekdays[day]}, ${pad(date.getUTCDate())}/${pad(date.getUTCMonth()+1)}${full?`/${date.getUTCFullYear()}`:''}`;
  }
  function availability(day,hour) {
    if ((day===1 && hour===10) || (day===6 && hour>=12 && hour<=14)) return {kind:'closed',label:'Đóng cửa',count:0};
    if ((day===0 && hour===8) || (day===2 && hour===11) || (day===4 && hour===18)) return {kind:'full',label:'Đã đầy',count:0};
    if ((day===4 && hour===7) || (day===3 && hour===17)) return {kind:'low',label:'Còn 1 chỗ',count:1};
    const count = Math.max(2,room().capacity-((day*3+hour)%5));
    return {kind:'open',label:`${count} chỗ`,count};
  }
  function isValid(candidate) {
    return candidate && candidate.day>=0 && candidate.day<7 && candidate.start>=firstHour && candidate.end<=lastHour && candidate.end>candidate.start &&
      Array.from({length:candidate.end-candidate.start},(_,i)=>availability(candidate.day,candidate.start+i).count>0).every(Boolean);
  }
  const rangeText = selection => selection?`${clock(selection.start)} – ${clock(selection.end)}`:'Chưa chọn giờ';
  function notify(text) { $('#toast').textContent=text; $('#toast').hidden=false; clearTimeout(toastTimer); toastTimer=setTimeout(()=>$('#toast').hidden=true,3400); }
  function error(text='') { $('#range-error').textContent=text; $('#range-error').hidden=!text; }
  function broadcast() { if(window.parent!==window) window.parent.postMessage({type:'u0-state',screen:state.screen==='success'?'booking':state.screen,persona:state.persona,scenario:state.scenario},'*'); }
  function renderRooms() {
    const visible = rooms.filter(item=>(state.filter==='all'||item.id===state.filter) && `${item.name} ${item.type} ${item.description}`.toLocaleLowerCase('vi').includes(state.query.toLocaleLowerCase('vi')));
    $('#room-cards').innerHTML = visible.map(item=>`<article class="room-card"><div class="room-card-img"><a href="#booking" data-room="${item.id}"><img src="${photoRoot+item.photo}" alt="Ảnh minh họa ${item.name}" loading="lazy"></a><span class="photo-tag">Ảnh minh họa</span></div><div class="room-card-body"><div class="room-card-topline"><span class="room-type">${item.type}</span><span>${item.capacity} chỗ / ca</span></div><h3>${item.name}</h3><p>${item.description}</p><div class="room-card-bottom"><span>Ca tập 60 phút</span><a href="#booking" data-room="${item.id}">Xem lịch <span aria-hidden="true">↗</span></a></div></div></article>`).join('');
    $('#empty-search').hidden=visible.length>0;
    $$('[data-filter]').forEach(button=>{const active=button.dataset.filter===state.filter;button.classList.toggle('active',active);button.setAttribute('aria-pressed',active);});
  }
  function setSelection(candidate, announce=true) {
    if (!isValid(candidate)) { error('Khoảng giờ đi qua ca đã đầy hoặc đóng cửa. Lựa chọn trước đó được giữ nguyên.'); return false; }
    state.selection={...candidate}; state.activeDay=candidate.day; state.trainer=null; error(); drawSelection(); renderDays(); renderTraining(); renderSummary(); syncAlternative();
    if(announce) notify(`Đã chọn ${dateLabel(candidate.day)} · ${rangeText(candidate)} · ${candidate.end-candidate.start} ca.`);
    return true;
  }
  function visibleDays() { return state.view==='day'?[state.activeDay]:[0,1,2,3,4,5,6]; }
  function renderCalendar() {
    const dates=visibleDays();
    $('#calendar-grid').classList.toggle('is-day',state.view==='day');
    $('#calendar-grid').innerHTML=`<div class="time-corner">Giờ VN</div>`+
      dates.map((day,index)=>`<div class="day-header ${day===state.activeDay?'is-active':''}" data-header="${day}" style="grid-column:${index+2};grid-row:1"><span>${shortDays[day]}</span><strong>${pad(dateFor(day).getUTCDate())}</strong></div>`).join('')+
      `<div class="time-axis">${Array.from({length:lastHour-firstHour},(_,i)=>`<div>${clock(firstHour+i)}</div>`).join('')}</div>`+
      dates.map((day,index)=>`<div class="day-column" data-column="${day}" style="grid-column:${index+2}">${Array.from({length:lastHour-firstHour},(_,i)=>{const hour=firstHour+i;const slot=availability(day,hour);return `<button class="calendar-cell ${slot.kind}" data-day="${day}" data-hour="${hour}" aria-label="${dateLabel(day)} ${clock(hour)} đến ${clock(hour+1)}, ${slot.label}" aria-disabled="${slot.count===0}"><span>${slot.label}</span></button>`;}).join('')}</div>`).join('');
    drawSelection(); renderDays(); updateMode();
  }
  function drawSelection() {
    $$('.selection-block').forEach(block=>block.remove());
    const selection=draft||state.selection;
    if(!selection)return;
    const column=$(`[data-column="${selection.day}"]`); if(!column)return;
    const block=document.createElement('div');
    const row=parseFloat(getComputedStyle(document.documentElement).getPropertyValue('--row-height'));
    block.className=`selection-block${draft?' draft':''}${isValid(selection)?'':' invalid'}`;
    block.style.top=`${(selection.start-firstHour)*row}px`;block.style.height=`${(selection.end-selection.start)*row}px`;
    block.dataset.selectedDay=selection.day;
    block.setAttribute('role','group');block.setAttribute('aria-label',`Khoảng đã chọn ${rangeText(selection)}`);
    block.innerHTML=`<button type="button" class="selection-handle top" data-handle="start" aria-label="Chỉnh giờ bắt đầu, dùng mũi tên lên xuống"></button><strong>${rangeText(selection)}</strong><small>${selection.end-selection.start} ca · ${(selection.end-selection.start)*60} phút</small><button type="button" class="selection-handle bottom" data-handle="end" aria-label="Chỉnh giờ kết thúc, dùng mũi tên lên xuống"></button>`;
    column.appendChild(block);
    $$('[data-header]').forEach(header=>header.classList.toggle('is-active',Number(header.dataset.header)===selection.day));
  }
  function renderDays() {
    const start=dateFor(0),end=dateFor(6);
    $('#week-label').textContent=`${pad(start.getUTCDate())} – ${pad(end.getUTCDate())}/${pad(end.getUTCMonth()+1)}/${end.getUTCFullYear()}`;
    $('#day-jump').innerHTML=shortDays.map((name,day)=>`<button data-jump="${day}" class="${day===state.activeDay?'active':''}" aria-pressed="${day===state.activeDay}" aria-label="Xem ${dateLabel(day)}"><span>${name}</span><strong>${pad(dateFor(day).getUTCDate())}</strong></button>`).join('');
  }
  function updateMode() {
    $('#calendar-scroll').classList.toggle('mode-select',state.mode==='select');
    $$('[data-mode]').forEach(button=>{const active=button.dataset.mode===state.mode;button.classList.toggle('active',active);button.setAttribute('aria-pressed',active);});
    $$('[data-view]').forEach(button=>{const active=button.dataset.view===state.view;button.classList.toggle('active',active);button.setAttribute('aria-pressed',active);});
    $('#calendar-tip').textContent=state.mode==='view'?'Vuốt ngang trong lịch để xem đủ 7 ngày, cuộn dọc để xem giờ. Chuyển sang Chọn giờ khi muốn kéo chọn.':'Kéo chọn trong một ngày. Kéo hai đầu để chỉnh; kéo giữa khối để đổi giờ. Hoặc chạm ca đầu, rồi ca cuối.';
  }
  function renderTraining() {
    const guest=state.persona==='guest';
    $('#trainer-hint').textContent=!state.selection?'Chọn khoảng giờ để xem HLV phù hợp.':guest?'Đăng nhập ở bước tiếp tục để kiểm tra HLV phù hợp.':'Các HLV mẫu bên dưới phù hợp với toàn khoảng giờ bạn chọn.';
    const self=`<button class="training-option ${state.training==='self'?'selected':''}" data-training="self" aria-pressed="${state.training==='self'}"><span class="avatar" aria-hidden="true"><svg viewBox="0 0 24 24"><path d="M4 8v8m4-11v14m8-14v14m4-11v8M8 12h8"/></svg></span><span><strong>Tự tập</strong><small>Xác nhận ngay khi còn chỗ</small></span><span class="radio-mark" aria-hidden="true"></span></button>`;
    const pt=guest?`<button class="training-option ${state.training==='pt'?'selected':''}" data-training="pt" aria-pressed="${state.training==='pt'}"><span class="avatar avatar-blue" aria-hidden="true">PT</span><span><strong>Tập cùng HLV</strong><small>Kiểm tra HLV sau đăng nhập</small></span><span class="radio-mark" aria-hidden="true"></span></button>`:trainers.map(item=>`<button class="training-option ${state.trainer===item.id?'selected':''}" data-trainer="${item.id}" aria-pressed="${state.trainer===item.id}"><span class="avatar ${item.id==='cardio'?'avatar-blue':''}" aria-hidden="true">${item.initials}</span><span><strong>${item.name}</strong><small>${item.specialty}</small></span><span class="radio-mark" aria-hidden="true"></span></button>`).join('');
    $('#training-options').innerHTML=self+pt;
  }
  function trainingLabel() {return state.training==='self'?'Tự tập':state.trainer?trainers.find(item=>item.id===state.trainer).name:'Cùng HLV · Chưa chọn';}
  function renderSummary() {
    const selection=draft||state.selection;
    const guest=state.persona==='guest', expired=state.persona==='expired';
    const count=selection?selection.end-selection.start:0;
    $('#summary-room').textContent=room().name;
    $('#summary-date').textContent=selection?dateLabel(selection.day,true):'Chọn ngày trên lịch';
    $('#summary-time').textContent=rangeText(selection);
    $('#summary-duration').textContent=selection?`${count} ca liền kề · ${count*60} phút`:'Chưa có khoảng giờ được chọn';
    $('#summary-training').textContent=trainingLabel();
    $('#summary-status').textContent=expired?'Thẻ đã hết hạn':selection?'Còn chỗ':'Chưa chọn';
    $('#summary-note').textContent=expired?'Liên hệ lễ tân để gia hạn thẻ trước khi đặt lịch.':guest?'Lựa chọn chưa giữ chỗ. Đăng nhập để tiếp tục đặt lịch.':state.training==='pt'?'Yêu cầu được giữ trong khi chờ PT phản hồi, tối đa 30 phút và trước khi ca bắt đầu.':'Lịch tự tập được xác nhận ngay khi còn chỗ. Có thể hủy trước giờ tập ít nhất 2 tiếng.';
    $('#continue-booking').innerHTML=`${guest?'Tiếp tục đặt lịch':'Xem lại đặt lịch'} <span aria-hidden="true">→</span>`;
    $('#summary-footnote').textContent='Bạn sẽ xem lại trước khi xác nhận.';
    const disabled=!state.selection||expired||state.scenario!=='normal';
    $('#continue-booking').disabled=disabled;$('#mobile-continue').disabled=disabled;
    $('#mobile-continue').innerHTML=`${guest?'Tiếp tục':'Xem lại'} <span aria-hidden="true">→</span>`;
    $('#mobile-time').textContent=rangeText(selection);
    $('#mobile-date').textContent=selection?`${dateLabel(selection.day)} · ${count} ca`:'Chọn một khoảng giờ trên lịch';
    $('#eligibility-notice').hidden=!expired;
    $('#account-button').textContent=guest?'Đăng nhập':expired?'Thẻ hết hạn':'Hội viên mẫu';
  }
  function renderFeedback() {
    const normal=state.scenario==='normal';
    $('#calendar-feedback').hidden=normal;$('#calendar-scroll').hidden=!normal;
    $('.time-alternative').hidden=!normal;$('#calendar-tip').hidden=!normal;
    if(normal)return;
    const content={
      loading:'<div class="skeleton-grid" aria-hidden="true">'+Array.from({length:7},()=>'<span></span>').join('')+'</div><p>Đang tải các ca còn chỗ…</p>',
      empty:'<h3>Tuần này chưa có ca tập.</h3><p>Bạn có thể xem tuần kế tiếp hoặc chọn phòng khác để tìm giờ phù hợp.</p><button class="button button-small" data-next-available>Xem tuần kế tiếp</button>',
      error:'<h3>Chưa tải được lịch.</h3><p>Kiểm tra kết nối rồi thử lại. Khoảng giờ bạn vừa chọn vẫn được giữ trong bản nháp.</p><button class="button button-small" data-retry>Thử lại</button>',
    };
    $('#calendar-feedback').innerHTML=content[state.scenario];
    $('#summary-note').textContent=state.scenario==='loading'?'Đang kiểm tra lịch trống. Bạn có thể tiếp tục khi tải xong.':state.scenario==='empty'?'Chọn tuần khác hoặc đổi phòng để xem các ca nhận đặt.':'Thử tải lại lịch trước khi tiếp tục đặt.';
  }
  function syncAlternative() {
    const selection=state.selection||{day:state.activeDay,start:7,end:9};
    $('#alt-day').innerHTML=weekdays.map((_,i)=>`<option value="${i}" ${selection.day===i?'selected':''}>${shortDays[i]} ${pad(dateFor(i).getUTCDate())}/${pad(dateFor(i).getUTCMonth()+1)}</option>`).join('');
    $('#alt-start').innerHTML=Array.from({length:lastHour-firstHour},(_,i)=>`<option value="${firstHour+i}" ${firstHour+i===selection.start?'selected':''}>${clock(firstHour+i)}</option>`).join('');
    $('#alt-end').innerHTML=Array.from({length:lastHour-firstHour},(_,i)=>`<option value="${firstHour+i+1}" ${firstHour+i+1===selection.end?'selected':''}>${clock(firstHour+i+1)}</option>`).join('');
  }
  function renderBooking() {
    $('#booking-room-name').textContent=room().name;$('#booking-category').textContent=room().type;
    $('#booking-room-meta').textContent=`${room().capacity} chỗ mỗi ca · 60 phút mỗi ca · Giờ Việt Nam`;
    $('#booking-photo').src=photoRoot+room().photo;
    $('#room-switch').innerHTML=rooms.map(item=>`<option value="${item.id}" ${state.room===item.id?'selected':''}>${item.name}</option>`).join('');
    renderCalendar();renderTraining();renderSummary();renderFeedback();syncAlternative();
  }
  function showScreen(screen, anchor) {
    state.screen=screen;
    $('#home-view').hidden=screen!=='home';$('#booking-view').hidden=screen!=='booking';$('#success-view').hidden=screen!=='success';
    $('#mobile-summary').hidden=screen!=='booking';
    $('[data-nav="home"]').classList.toggle('active',screen==='home');
    $('.site-header nav').classList.remove('open');$('#menu-button').setAttribute('aria-expanded','false');
    if(screen==='booking') {renderBooking();requestAnimationFrame(()=>scrollDay(state.activeDay));}
    if(anchor) requestAnimationFrame(()=>$(anchor)?.scrollIntoView({behavior:'auto',block:'start'}));else window.scrollTo({top:0,behavior:'auto'});
    broadcast();
  }
  function route() {
    const hash=location.hash.slice(1)||'home';
    if(hash==='booking')showScreen('booking');else if(['rooms','trainers','guide'].includes(hash))showScreen('home','#'+hash);else showScreen('home');
  }
  function scrollDay(day) {
    if(state.view==='day')return;
    const column=$(`[data-column="${day}"]`);if(!column)return;
    const scroll=$('#calendar-scroll');const dayWidth=column.getBoundingClientRect().width;
    if(scroll.clientWidth<$('#calendar-grid').scrollWidth) scroll.scrollTo({left:day*dayWidth,behavior:'auto'});
  }
  function changeWeek(delta) {
    const hadSelection=!!state.selection;
    state.week=delta===0?0:state.week+delta;state.selection=null;state.trainer=null;tapAnchor=null;draft=null;
    renderBooking();error();
    notify(hadSelection?'Đã chuyển tuần và xóa lựa chọn cũ. Chọn khoảng giờ mới.':'Đã chuyển tuần. Chọn khoảng giờ phù hợp.');
  }
  function changePersona(persona) {
    if(!['guest','member','expired'].includes(persona))return;
    state.persona=persona;state.trainer=null;
    renderTraining();renderSummary();broadcast();
  }
  function openLogin() {
    $('#toast').hidden=true;
    $('#login-recap').innerHTML=state.selection?`<strong>${room().name}</strong><br>${dateLabel(state.selection.day)} · ${rangeText(state.selection)}`:'Khám phá phòng và giờ tập phù hợp với bạn.';
    $('#login-dialog').showModal();
  }
  function continueBooking() {
    if(!state.selection||state.scenario!=='normal')return;
    if(state.persona==='guest'){openLogin();return;}
    if(state.persona==='expired'){notify('Thẻ đã hết hạn. Liên hệ lễ tân để gia hạn.');return;}
    if(state.training==='pt'&&!state.trainer){notify('Chọn HLV phù hợp trước khi xem lại.');$('.training-panel').scrollIntoView({block:'center'});return;}
    if(!isValid(state.selection)){error('Lịch đã thay đổi. Chọn lại khoảng giờ còn chỗ.');return;}
    const details=[['Phòng tập',room().name],['Ngày tập',dateLabel(state.selection.day,true)],['Khoảng giờ',rangeText(state.selection)],['Thời lượng',`${state.selection.end-state.selection.start} ca · ${(state.selection.end-state.selection.start)*60} phút`],['Hình thức',trainingLabel()]];
    $('#review-rows').innerHTML=details.map(([name,value])=>`<div class="review-row"><span>${name}</span><strong>${value}</strong></div>`).join('');
    $('#review-note').textContent=state.training==='self'?'Lịch tự tập sẽ được xác nhận ngay khi còn chỗ. Bạn có thể hủy trước giờ tập ít nhất 2 tiếng.':'Yêu cầu cần được HLV xác nhận trong tối đa 30 phút và trước khi ca bắt đầu.';
    $('#toast').hidden=true;$('#review-dialog').showModal();
  }
  function hit(event) {
    const columns=$$('.day-column');const column=columns.find(item=>{const rect=item.getBoundingClientRect();return event.clientX>=rect.left&&event.clientX<rect.right;});
    if(!column)return null;
    const row=parseFloat(getComputedStyle(document.documentElement).getPropertyValue('--row-height'));
    const hour=firstHour+Math.floor((event.clientY-column.getBoundingClientRect().top)/row);
    return {day:Number(column.dataset.column),hour:Math.max(firstHour,Math.min(lastHour-1,hour))};
  }
  function tapSelect(day,hour) {
    if(tapAnchor&&tapAnchor.day===day){const candidate={day,start:Math.min(tapAnchor.hour,hour),end:Math.max(tapAnchor.hour,hour)+1};if(setSelection(candidate))tapAnchor=null;}
    else if(setSelection({day,start:hour,end:hour+1})){tapAnchor={day,hour};}
  }
  const calendar=$('#calendar-scroll');
  calendar.addEventListener('pointerdown',event=>{
    if(state.mode!=='select'||event.button!==0)return;
    const point=hit(event);if(!point)return;
    const selection=event.target.closest('.selection-block');const cell=event.target.closest('.calendar-cell');if(!selection&&!cell)return;
    event.preventDefault();
    const handle=event.target.closest('[data-handle]');
    drag={kind:handle?handle.dataset.handle:selection?'move':'create',day:point.day,anchor:point.hour,base:state.selection?{...state.selection}:null,x:event.clientX,y:event.clientY,moved:false};
    if(drag.kind==='create')draft={day:point.day,start:point.hour,end:point.hour+1};else draft={...state.selection};
    calendar.setPointerCapture(event.pointerId);drawSelection();
  });
  calendar.addEventListener('pointermove',event=>{
    if(!drag)return;
    const point=hit(event);if(!point||point.day!==drag.day)return;
    if(Math.abs(event.clientY-drag.y)>4||Math.abs(event.clientX-drag.x)>4)drag.moved=true;
    const base=drag.base;
    if(drag.kind==='create')draft={day:drag.day,start:Math.min(drag.anchor,point.hour),end:Math.max(drag.anchor,point.hour)+1};
    else if(drag.kind==='start')draft={...base,start:Math.min(point.hour,base.end-1)};
    else if(drag.kind==='end')draft={...base,end:Math.max(point.hour+1,base.start+1)};
    else {const length=base.end-base.start;const start=Math.max(firstHour,Math.min(lastHour-length,base.start+point.hour-drag.anchor));draft={day:drag.day,start,end:start+length};}
    drawSelection();
    const rect=calendar.getBoundingClientRect();if(event.clientY>rect.bottom-35)calendar.scrollTop+=7;else if(event.clientY<rect.top+75)calendar.scrollTop-=7;
  });
  calendar.addEventListener('pointerup',event=>{
    if(!drag)return;
    const active=drag,candidate=draft;const point=hit(event);draft=null;drag=null;
    if(calendar.hasPointerCapture(event.pointerId))calendar.releasePointerCapture(event.pointerId);
    const bounds=calendar.getBoundingClientRect();
    if(!point||event.clientY<bounds.top||event.clientY>bounds.bottom){drawSelection();return;}
    if(active.kind==='create'&&!active.moved)tapSelect(active.day,active.anchor);
    else if(active.moved){tapAnchor=null;setSelection(candidate);}
    drawSelection();
  });
  calendar.addEventListener('pointercancel',()=>{drag=null;draft=null;drawSelection();});
  calendar.addEventListener('keydown',event=>{
    if(event.key==='Escape'){drag=null;draft=null;tapAnchor=null;error();drawSelection();return;}
    const handle=event.target.closest('[data-handle]');
    if(handle&&['ArrowUp','ArrowDown'].includes(event.key)&&state.selection){event.preventDefault();const candidate={...state.selection};candidate[handle.dataset.handle]+=event.key==='ArrowUp'?-1:1;const which=handle.dataset.handle;if(setSelection(candidate))$(`[data-handle="${which}"]`)?.focus();return;}
    const cell=event.target.closest('.calendar-cell');if(!cell)return;
    const day=Number(cell.dataset.day),hour=Number(cell.dataset.hour);
    if(['Enter',' '].includes(event.key)){event.preventDefault();tapSelect(day,hour);return;}
    const directions={ArrowUp:[0,-1],ArrowDown:[0,1],ArrowLeft:[-1,0],ArrowRight:[1,0]};const direction=directions[event.key];
    if(direction){event.preventDefault();const target=$(`[data-day="${day+direction[0]}"][data-hour="${hour+direction[1]}"]`);target?.focus();}
  });
  document.addEventListener('click',event=>{
    if(event.target.closest('[data-retry]')){state.scenario='normal';renderBooking();broadcast();notify('Đã tải lại lịch mẫu.');}
    if(event.target.closest('[data-next-available]')){state.scenario='normal';changeWeek(1);broadcast();}
    const filter=event.target.closest('[data-filter]');if(filter){state.filter=filter.dataset.filter;renderRooms();}
    const roomLink=event.target.closest('[data-room]');if(roomLink){if(state.room!==roomLink.dataset.room){state.selection=null;notify('Đã đổi phòng. Chọn khoảng giờ mới.');}state.room=roomLink.dataset.room;state.trainer=null;error();if(location.hash==='#booking')showScreen('booking');}
    const view=event.target.closest('[data-view]');if(view){state.view=view.dataset.view;renderCalendar();scrollDay(state.activeDay);}
    const mode=event.target.closest('[data-mode]');if(mode){state.mode=mode.dataset.mode;tapAnchor=null;updateMode();}
    const jump=event.target.closest('[data-jump]');if(jump){state.activeDay=Number(jump.dataset.jump);if(state.view==='day')renderCalendar();else renderDays();scrollDay(state.activeDay);}
    const training=event.target.closest('[data-training]');if(training){state.training=training.dataset.training;state.trainer=null;renderTraining();renderSummary();}
    const trainer=event.target.closest('[data-trainer]');if(trainer){state.training='pt';state.trainer=trainer.dataset.trainer;renderTraining();renderSummary();}
    const close=event.target.closest('[data-close]');if(close)close.closest('dialog').close();
  });
  $('#room-search').addEventListener('input',event=>{state.query=event.target.value;renderRooms();});
  $('#room-switch').addEventListener('change',event=>{state.room=event.target.value;state.selection=null;state.trainer=null;tapAnchor=null;renderBooking();error();notify('Đã đổi phòng và xóa lựa chọn cũ. Chọn khoảng giờ mới.');});
  $('#menu-button').addEventListener('click',()=>{const open=$('.site-header nav').classList.toggle('open');$('#menu-button').setAttribute('aria-expanded',open);});
  $('#prev-week').addEventListener('click',()=>changeWeek(-1));$('#next-week').addEventListener('click',()=>changeWeek(1));$('#sample-week').addEventListener('click',()=>changeWeek(0));
  $('#clear-selection').addEventListener('click',()=>{state.selection=null;state.trainer=null;tapAnchor=null;draft=null;drawSelection();renderTraining();renderSummary();syncAlternative();error();notify('Đã xóa lựa chọn.');});
  $('#apply-times').addEventListener('click',()=>{if(setSelection({day:Number($('#alt-day').value),start:Number($('#alt-start').value),end:Number($('#alt-end').value)})){if(state.view==='day')renderCalendar();scrollDay(state.activeDay);}});
  $('#account-button').addEventListener('click',()=>state.persona==='guest'?openLogin():notify(state.persona==='expired'?'Thẻ mẫu đã hết hạn. Liên hệ lễ tân để gia hạn.':'Bạn đang dùng tài khoản Hội viên mẫu trong bản thiết kế.'));
  $('#continue-booking').addEventListener('click',continueBooking);$('#mobile-continue').addEventListener('click',continueBooking);
  $('#simulate-login').addEventListener('click',()=>{$('#login-dialog').close();changePersona('member');notify('Đã dùng hội viên mẫu. Lựa chọn được giữ nguyên; chưa tạo lịch.');if(state.screen==='booking'&&state.training==='pt')$('.training-panel').scrollIntoView({block:'center'});});
  $('#confirm-booking').addEventListener('click',()=>{
    if(!state.selection)return;$('#review-dialog').close();
    $('#success-room').textContent=room().name;$('#success-date').textContent=dateLabel(state.selection.day,true);$('#success-time').textContent=rangeText(state.selection);$('#success-training').textContent=trainingLabel();
    $('#success-status').textContent=state.training==='self'?'Lịch tự tập mẫu đã được xác nhận.':'Yêu cầu mẫu đang chờ HLV phản hồi.';showScreen('success');
  });
  $('#book-again').addEventListener('click',()=>{state.selection=null;state.trainer=null;showScreen('booking');});
  window.addEventListener('hashchange',route);
  window.addEventListener('keydown',event=>{if(event.key==='Escape'&&drag){drag=null;draft=null;tapAnchor=null;drawSelection();error();}});
  window.addEventListener('message',event=>{
    if(event.source!==window.parent||event.data?.type!=='u0-control')return;
    if(event.data.persona)changePersona(event.data.persona);
    if(['normal','loading','empty','error'].includes(event.data.scenario)){state.scenario=event.data.scenario;showScreen('booking');}
    if(['home','booking'].includes(event.data.screen)){if(location.hash==='#'+event.data.screen)showScreen(event.data.screen);else location.hash=event.data.screen;}
  });
  matchMedia('(max-width:650px)').addEventListener('change',event=>{state.mode=event.matches?'view':'select';updateMode();drawSelection();});
  if(typeof ResizeObserver!=='undefined')new ResizeObserver(()=>{if(state.screen==='booking'&&!drag)scrollDay(state.selection?.day??state.activeDay);}).observe(calendar);
  renderRooms();renderBooking();route();
})();
