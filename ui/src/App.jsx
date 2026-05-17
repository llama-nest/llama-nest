import React, {useEffect, useState} from 'react';
import {createRoot} from 'react-dom/client';
import logo from '../rsc/llama-nest.jpeg'
import {Database, Search, RefreshCcw, MessageSquare, Shield} from 'lucide-react';
import './style.css';

const API = 'http://localhost:8787';

function App(){
  const [status,setStatus]=useState({});
  const [sessions,setSessions]=useState([]);
  const [selected,setSelected]=useState(null);
  const [messages,setMessages]=useState([]);
  const [q,setQ]=useState('');
  const [results,setResults]=useState([]);
  const [brief,setBrief]=useState('');

  async function refresh(){
    setStatus(await fetch(API+'/api/status').then(r=>r.json()).catch(()=>({error:'API unavailable'})));
    setSessions(await fetch(API+'/api/sessions').then(r=>r.json()).catch(()=>[]));
  }
  async function openSession(s){ setSelected(s); setMessages(await fetch(API+'/api/messages?session_id='+s.id).then(r=>r.json())); }
  async function search(){ setResults(await fetch(API+'/api/search?q='+encodeURIComponent(q)).then(r=>r.json())); }
  async function catchUp(){ const x=await fetch(API+'/api/catch-up').then(r=>r.json()); setBrief(x.brief); }
  useEffect(()=>{refresh()},[]);

  return <div className="app">
    <aside>
      <div className="brand"> <img src={logo} alt="llama-nest" className="logo" width="50" height="50"/> <div><h1>Llama Nest</h1> <p>local AI memory</p></div> </div>
      <button onClick={refresh}><RefreshCcw size={16}/> Refresh</button>
      <button onClick={catchUp}><MessageSquare size={16}/> Catch up</button>
      <div className="stat"><Database size={16}/> {status.sessions ?? 0} sessions · {status.messages ?? 0} messages</div>
      <div className="privacy"><Shield size={16}/> Local-first v0. No telemetry.</div>
    </aside>
    <main>
      <section className="hero">
        <h2>Inspectable memory for Ollama.</h2>
        <p>Run Ollama traffic through <code>localhost:11435</code>. llama-nest captures context locally so you can inspect, search, and generate catch-up briefs.</p>
      </section>
      <section className="grid">
        <div className="card">
          <h3>Sessions</h3>
          {sessions.length===0 && <p className="muted">No captured sessions yet.</p>}
          {sessions.map(s=><div className={'row '+(selected?.id===s.id?'active':'')} key={s.id} onClick={()=>openSession(s)}><b>{s.title}</b><span>{s.model || 'unknown model'}</span></div>)}
        </div>
        <div className="card">
          <h3>{selected ? selected.title : 'Messages'}</h3>
          {messages.map(m=><div className="msg" key={m.id}><span>{m.role}</span><p>{m.content}</p></div>)}
          {!selected && <p className="muted">Select a session.</p>}
        </div>
      </section>
      <section className="grid">
        <div className="card">
          <h3><Search size={16}/> Search</h3>
          <div className="search"><input value={q} onChange={e=>setQ(e.target.value)} placeholder="Search local context" onKeyDown={e=>{if(e.key==='Enter')search()}}/><button onClick={search}>Search</button></div>
          {results.map(r=><div className="msg" key={r.id}><span>{r.role} · session {r.session_id}</span><p>{r.content}</p></div>)}
        </div>
        <div className="card">
          <h3>Catch-up brief</h3>
          <pre>{brief || 'Click “Catch up” to create a recent-memory brief.'}</pre>
        </div>
      </section>
    </main>
  </div>
}

createRoot(document.getElementById('root')).render(<App/>);
